package implementations

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/pkg/async/cloudevent/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/async/notifications/implementations"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	dataInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/data/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const (
	cloudEventSource     = "https://github.com/flyteorg/flyte/flyteadmin"
	cloudEventTypePrefix = "com.flyte.resource"
	jsonSchemaURLKey     = "jsonschemaurl"
	jsonSchemaURL        = "https://github.com/flyteorg/flyteidl/blob/v0.24.14/jsonschema/workflow_execution.json"
)

// Publisher This event publisher acts to asynchronously publish workflow execution events.
type Publisher struct {
	sender        interfaces.Sender
	systemMetrics implementations.EventPublisherSystemMetrics
	events        sets.String
}

func (p *Publisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	if !p.shouldPublishEvent(notificationType) {
		return nil
	}
	p.systemMetrics.PublishTotal.Inc()
	logger.Debugf(ctx, "Publishing the following message [%+v]", msg)

	var executionID string
	var phase string
	var eventTime time.Time

	switch msgType := msg.(type) {
	case *admin.WorkflowExecutionEventRequest:
		e := msgType.GetEvent()
		executionID = e.GetExecutionId().String()
		phase = e.GetPhase().String()
		eventTime = e.GetOccurredAt().AsTime()
	case *admin.TaskExecutionEventRequest:
		e := msgType.GetEvent()
		executionID = e.GetTaskId().String()
		phase = e.GetPhase().String()
		eventTime = e.GetOccurredAt().AsTime()
	case *admin.NodeExecutionEventRequest:
		e := msgType.GetEvent()
		executionID = msgType.GetEvent().GetId().String()
		phase = e.GetPhase().String()
		eventTime = e.GetOccurredAt().AsTime()
	default:
		return fmt.Errorf("unsupported event types [%+v]", reflect.TypeOf(msg))
	}

	event := cloudevents.NewEvent()
	// CloudEvent specification: https://github.com/cloudevents/spec/blob/v1.0/spec.md#required-attributes
	event.SetType(fmt.Sprintf("%v.%v", cloudEventTypePrefix, notificationType))
	event.SetSource(cloudEventSource)
	event.SetID(fmt.Sprintf("%v.%v", executionID, phase))
	event.SetTime(eventTime)
	event.SetExtension(jsonSchemaURLKey, jsonSchemaURL)

	// Explicitly jsonpb marshal the proto. Otherwise, event.SetData will use json.Marshal which doesn't work well
	// with proto oneof fields.
	marshaler := jsonpb.Marshaler{}
	buf := bytes.NewBuffer([]byte{})
	err := marshaler.Marshal(buf, msg)
	if err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to jsonpb marshal [%v] with error: %v", msg, err)
		return fmt.Errorf("failed to jsonpb marshal [%v] with error: %w", msg, err)
	}

	if err := event.SetData(cloudevents.ApplicationJSON, buf.Bytes()); err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to encode message [%v] with error: %v", msg, err)
		return err
	}

	if err := p.sender.Send(ctx, notificationType, event); err != nil {
		p.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to send message [%v] with error: %v", msg, err)
		return err
	}
	p.systemMetrics.PublishSuccess.Inc()
	return nil
}

func (p *Publisher) shouldPublishEvent(notificationType string) bool {
	return p.events.Has(notificationType)
}

type CloudEventWrappedPublisher struct {
	db                   repositoryInterfaces.Repository
	sender               interfaces.Sender
	systemMetrics        implementations.EventPublisherSystemMetrics
	storageClient        *storage.DataStore
	urlData              dataInterfaces.RemoteURLInterface
	remoteDataConfig     runtimeInterfaces.RemoteDataConfig
	eventPublisherConfig runtimeInterfaces.EventsPublisherConfig
}

func (c *CloudEventWrappedPublisher) TransformWorkflowExecutionEvent(ctx context.Context, rawEvent *event.WorkflowExecutionEvent) (*event.CloudEventWorkflowExecution, error) {

	// Basic error checking
	if rawEvent == nil {
		return nil, fmt.Errorf("nothing to publish, WorkflowExecution event is nil")
	}
	if rawEvent.GetExecutionId() == nil {
		logger.Warningf(ctx, "nil execution id in event [%+v]", rawEvent)
		return nil, fmt.Errorf("nil execution id in event [%+v]", rawEvent)
	}

	// For now, don't append any additional information unless succeeded or otherwise configured
	if rawEvent.GetPhase() != core.WorkflowExecution_SUCCEEDED && !c.eventPublisherConfig.EnrichAllWorkflowEventTypes {
		return &event.CloudEventWorkflowExecution{
			RawEvent: rawEvent,
		}, nil
	}

	// TODO: Make this one call to the DB instead of two.
	executionModel, err := c.db.ExecutionRepo().Get(ctx, repositoryInterfaces.Identifier{
		Project: rawEvent.GetExecutionId().GetProject(),
		Domain:  rawEvent.GetExecutionId().GetDomain(),
		Name:    rawEvent.GetExecutionId().GetName(),
	})
	if err != nil {
		logger.Warningf(ctx, "couldn't find execution [%+v] for cloud event processing", rawEvent.GetExecutionId())
		return nil, err
	}
	ex, err := transformers.FromExecutionModel(ctx, executionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Warningf(ctx, "couldn't transform execution [%+v] for cloud event processing", rawEvent.GetExecutionId())
		return nil, err
	}
	if ex.GetClosure().GetWorkflowId() == nil {
		logger.Warningf(ctx, "workflow id is nil for execution [%+v]", ex)
		return nil, fmt.Errorf("workflow id is nil for execution [%+v]", ex)
	}
	workflowModel, err := c.db.WorkflowRepo().Get(ctx, repositoryInterfaces.Identifier{
		Project: ex.GetClosure().GetWorkflowId().GetProject(),
		Domain:  ex.GetClosure().GetWorkflowId().GetDomain(),
		Name:    ex.GetClosure().GetWorkflowId().GetName(),
		Version: ex.GetClosure().GetWorkflowId().GetVersion(),
	})
	if err != nil {
		logger.Warningf(ctx, "couldn't find workflow [%+v] for cloud event processing", ex.GetClosure().GetWorkflowId())
		return nil, err
	}
	var workflowInterface core.TypedInterface
	if len(workflowModel.TypedInterface) > 0 {
		err = proto.Unmarshal(workflowModel.TypedInterface, &workflowInterface)
		if err != nil {
			return nil, fmt.Errorf(
				"artifact eventing - failed to unmarshal TypedInterface for workflow [%+v] with err: %v",
				workflowModel.ID, err)
		}
	}

	// The spec is used to retrieve metadata fields
	spec := &admin.ExecutionSpec{}
	err = proto.Unmarshal(executionModel.Spec, spec)
	if err != nil {
		fmt.Printf("there was an error with spec %v %v", err, executionModel.Spec)
	}

	return &event.CloudEventWorkflowExecution{
		RawEvent:           rawEvent,
		OutputInterface:    &workflowInterface,
		ArtifactIds:        spec.GetMetadata().GetArtifactIds(),
		ReferenceExecution: spec.GetMetadata().GetReferenceExecution(),
		Principal:          spec.GetMetadata().GetPrincipal(),
		LaunchPlanId:       spec.GetLaunchPlan(),
		Labels:             spec.GetLabels().GetValues(),
	}, nil
}

func getNodeExecutionContext(ctx context.Context, identifier *core.NodeExecutionIdentifier) context.Context {
	ctx = contextutils.WithProjectDomain(ctx, identifier.GetExecutionId().GetProject(), identifier.GetExecutionId().GetDomain())
	ctx = contextutils.WithExecutionID(ctx, identifier.GetExecutionId().GetName())
	return contextutils.WithNodeID(ctx, identifier.GetNodeId())
}

// This is a rough copy of the ListTaskExecutions function in TaskExecutionManager. It can be deprecated once we move the processing out of Admin itself.
// Just return the highest retry attempt.
func (c *CloudEventWrappedPublisher) getLatestTaskExecutions(ctx context.Context, nodeExecutionID *core.NodeExecutionIdentifier) (*admin.TaskExecution, error) {
	ctx = getNodeExecutionContext(ctx, nodeExecutionID)

	identifierFilters, err := util.GetNodeExecutionIdentifierFilters(ctx, nodeExecutionID, common.TaskExecution)
	if err != nil {
		return nil, err
	}

	sort := admin.Sort{
		Key:       "retry_attempt",
		Direction: 0,
	}
	sortParameter, err := common.NewSortParameter(&sort, models.TaskExecutionColumns)
	if err != nil {
		return nil, err
	}

	output, err := c.db.TaskExecutionRepo().List(ctx, repositoryInterfaces.ListResourceInput{
		InlineFilters: identifierFilters,
		Offset:        0,
		Limit:         1,
		SortParameter: sortParameter,
	})
	if err != nil {
		return nil, err
	}
	if len(output.TaskExecutions) == 0 {
		logger.Debugf(ctx, "no task executions found for node exec id [%+v]", nodeExecutionID)
		return nil, nil
	}

	taskExecutionList, err := transformers.FromTaskExecutionModels(output.TaskExecutions, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Debugf(ctx, "failed to transform task execution models for node exec id [%+v] with err: %v", nodeExecutionID, err)
		return nil, err
	}

	return taskExecutionList[0], nil
}

func (c *CloudEventWrappedPublisher) TransformNodeExecutionEvent(ctx context.Context, rawEvent *event.NodeExecutionEvent) (*event.CloudEventNodeExecution, error) {
	if rawEvent == nil || rawEvent.GetId() == nil {
		return nil, fmt.Errorf("nothing to publish, NodeExecution event or ID is nil")
	}

	// Skip nodes unless they're succeeded and not start nodes
	if rawEvent.GetPhase() != core.NodeExecution_SUCCEEDED {
		return &event.CloudEventNodeExecution{
			RawEvent: rawEvent,
		}, nil
	} else if rawEvent.GetId().GetNodeId() == "start-node" {
		return &event.CloudEventNodeExecution{
			RawEvent: rawEvent,
		}, nil
	}
	// metric

	// This gets the parent workflow execution metadata
	executionModel, err := c.db.ExecutionRepo().Get(ctx, repositoryInterfaces.Identifier{
		Project: rawEvent.GetId().GetExecutionId().GetProject(),
		Domain:  rawEvent.GetId().GetExecutionId().GetDomain(),
		Name:    rawEvent.GetId().GetExecutionId().GetName(),
	})
	if err != nil {
		logger.Infof(ctx, "couldn't find execution [%+v] for cloud event processing", rawEvent.GetId().GetExecutionId())
		return nil, err
	}

	spec := &admin.ExecutionSpec{}
	err = proto.Unmarshal(executionModel.Spec, spec)
	if err != nil {
		fmt.Printf("there was an error with spec %v %v", err, executionModel.Spec)
	}

	// Fetch the latest task execution if any, and pull out the task interface, if applicable.
	// These are optional fields... if the node execution doesn't have a task execution then these will be empty.
	var taskExecID *core.TaskExecutionIdentifier
	var typedInterface *core.TypedInterface

	lte, err := c.getLatestTaskExecutions(ctx, rawEvent.GetId())
	if err != nil {
		logger.Errorf(ctx, "failed to get latest task execution for node exec id [%+v] with err: %v", rawEvent.GetId(), err)
		return nil, err
	}
	if lte != nil {
		taskModel, err := c.db.TaskRepo().Get(ctx, repositoryInterfaces.Identifier{
			Project: lte.GetId().GetTaskId().GetProject(),
			Domain:  lte.GetId().GetTaskId().GetDomain(),
			Name:    lte.GetId().GetTaskId().GetName(),
			Version: lte.GetId().GetTaskId().GetVersion(),
		})
		if err != nil {
			// TODO: metric this
			// metric
			logger.Debugf(ctx, "Failed to get task with task id [%+v] with err %v", lte.GetId().GetTaskId(), err)
			return nil, err
		}
		task, err := transformers.FromTaskModel(taskModel)
		if err != nil {
			logger.Debugf(ctx, "Failed to transform task model with err %v", err)
			return nil, err
		}
		typedInterface = task.GetClosure().GetCompiledTask().GetTemplate().GetInterface()
		taskExecID = lte.GetId()
	}

	return &event.CloudEventNodeExecution{
		RawEvent:        rawEvent,
		TaskExecId:      taskExecID,
		OutputInterface: typedInterface,
		ArtifactIds:     spec.GetMetadata().GetArtifactIds(),
		Principal:       spec.GetMetadata().GetPrincipal(),
		LaunchPlanId:    spec.GetLaunchPlan(),
		Labels:          spec.GetLabels().GetValues(),
	}, nil
}

func (c *CloudEventWrappedPublisher) TransformTaskExecutionEvent(ctx context.Context, rawEvent *event.TaskExecutionEvent) (*event.CloudEventTaskExecution, error) {

	if rawEvent == nil {
		return nil, fmt.Errorf("nothing to publish, TaskExecution event is nil")
	}

	executionModel, err := c.db.ExecutionRepo().Get(ctx, repositoryInterfaces.Identifier{
		Project: rawEvent.GetParentNodeExecutionId().GetExecutionId().GetProject(),
		Domain:  rawEvent.GetParentNodeExecutionId().GetExecutionId().GetDomain(),
		Name:    rawEvent.GetParentNodeExecutionId().GetExecutionId().GetName(),
	})
	if err != nil {
		logger.Warningf(ctx, "couldn't find execution [%+v] for cloud event processing", rawEvent.GetParentNodeExecutionId().GetExecutionId())
		return nil, err
	}
	ex, err := transformers.FromExecutionModel(ctx, executionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Warningf(ctx, "couldn't transform execution [%+v] for cloud event processing", rawEvent.GetParentNodeExecutionId().GetExecutionId())
		return nil, err
	}

	return &event.CloudEventTaskExecution{
		RawEvent: rawEvent,
		Labels:   ex.GetSpec().GetLabels().GetValues(),
	}, nil
}

func (c *CloudEventWrappedPublisher) Publish(ctx context.Context, notificationType string, msg proto.Message) error {
	c.systemMetrics.PublishTotal.Inc()
	logger.Debugf(ctx, "Publishing the following message [%+v]", msg)

	var err error
	var executionID string
	var phase string
	var eventTime time.Time
	var finalMsg proto.Message
	// this is a modified notification type. will be used for both event type and publishing topic.
	var topic string
	var eventID string
	var eventSource = cloudEventSource

	switch msgType := msg.(type) {
	case *admin.WorkflowExecutionEventRequest:
		topic = "cloudevents.WorkflowExecution"
		e := msgType.GetEvent()
		executionID = e.GetExecutionId().String()
		phase = e.GetPhase().String()
		eventTime = e.GetOccurredAt().AsTime()

		dummyNodeExecutionID := &core.NodeExecutionIdentifier{
			NodeId:      "end-node",
			ExecutionId: e.GetExecutionId(),
		}
		// This forms part of the key in the Artifact store,
		// but it should probably be entirely derived by that service instead.
		eventSource = common.FlyteURLKeyFromNodeExecutionID(dummyNodeExecutionID)
		finalMsg, err = c.TransformWorkflowExecutionEvent(ctx, e)
		if err != nil {
			logger.Errorf(ctx, "Failed to transform workflow execution event with error: %v", err)
			return err
		}
		eventID = fmt.Sprintf("%v.%v", executionID, phase)

	case *admin.TaskExecutionEventRequest:
		topic = "cloudevents.TaskExecution"
		e := msgType.GetEvent()
		executionID = e.GetTaskId().String()
		phase = e.GetPhase().String()
		eventTime = e.GetOccurredAt().AsTime()
		eventID = fmt.Sprintf("%v.%v", executionID, phase)

		if e.GetParentNodeExecutionId() == nil {
			return fmt.Errorf("parent node execution id is nil for task execution [%+v]", e)
		}
		eventSource = common.FlyteURLKeyFromNodeExecutionIDRetry(e.GetParentNodeExecutionId(),
			int(e.GetRetryAttempt()))
		finalMsg, err = c.TransformTaskExecutionEvent(ctx, e)
		if err != nil {
			logger.Errorf(ctx, "Failed to transform task execution event with error: %v", err)
			return err
		}
	case *admin.NodeExecutionEventRequest:
		topic = "cloudevents.NodeExecution"
		e := msgType.GetEvent()
		executionID = msgType.GetEvent().GetId().String()
		phase = e.GetPhase().String()
		eventTime = e.GetOccurredAt().AsTime()
		eventID = fmt.Sprintf("%v.%v", executionID, phase)
		eventSource = common.FlyteURLKeyFromNodeExecutionID(msgType.GetEvent().GetId())
		finalMsg, err = c.TransformNodeExecutionEvent(ctx, e)
		if err != nil {
			logger.Errorf(ctx, "Failed to transform node execution event with error: %v", err)
			return err
		}
	case *event.CloudEventExecutionStart:
		topic = "cloudevents.ExecutionStart"
		executionID = msgType.GetExecutionId().String()
		eventID = fmt.Sprintf("%v", executionID)
		eventTime = time.Now()
		// CloudEventExecutionStart don't have a nested event
		finalMsg = msgType
	default:
		return fmt.Errorf("unsupported event types [%+v]", reflect.TypeOf(msg))
	}

	// Explicitly jsonpb marshal the proto. Otherwise, event.SetData will use json.Marshal which doesn't work well
	// with proto oneof fields.
	marshaler := jsonpb.Marshaler{}
	buf := bytes.NewBuffer([]byte{})
	err = marshaler.Marshal(buf, finalMsg)
	if err != nil {
		c.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to jsonpb marshal [%v] with error: %v", msg, err)
		return fmt.Errorf("failed to jsonpb marshal [%v] with error: %w", msg, err)
	}

	cloudEvt := cloudevents.NewEvent()
	// CloudEvent specification: https://github.com/cloudevents/spec/blob/v1.0/spec.md#required-attributes
	cloudEvt.SetType(fmt.Sprintf("%v.%v", cloudEventTypePrefix, topic))
	// According to the spec, the combination of source and id should be unique.
	// Artifact service's uniqueness is project/domain/suffix. project/domain are available from the execution id.
	// so set the suffix as the source. Can ignore ID since Artifact will only listen to succeeded events.
	cloudEvt.SetSource(eventSource)
	cloudEvt.SetID(eventID)
	cloudEvt.SetTime(eventTime)
	// TODO: Fill this in after we can get auto-generation in buf.
	cloudEvt.SetExtension(jsonSchemaURLKey, "")

	if err := cloudEvt.SetData(cloudevents.ApplicationJSON, buf.Bytes()); err != nil {
		c.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to encode message [%v] with error: %v", msg, err)
		return err
	}

	if err := c.sender.Send(ctx, topic, cloudEvt); err != nil {
		c.systemMetrics.PublishError.Inc()
		logger.Errorf(ctx, "Failed to send message [%v] with error: %v", msg, err)
		return err
	}
	c.systemMetrics.PublishSuccess.Inc()
	return nil
}

func NewCloudEventsPublisher(sender interfaces.Sender, scope promutils.Scope, eventTypes []string) interfaces.Publisher {
	eventSet := sets.NewString()

	for _, eventType := range eventTypes {
		if eventType == implementations.AllTypes || eventType == implementations.AllTypesShort {
			for _, e := range implementations.SupportedEvents {
				eventSet = eventSet.Insert(e)
			}
			break
		}
		if e, found := implementations.SupportedEvents[eventType]; found {
			eventSet = eventSet.Insert(e)
		} else {
			panic(fmt.Errorf("unsupported event type [%s] in the config", eventType))
		}
	}

	return &Publisher{
		sender:        sender,
		systemMetrics: implementations.NewEventPublisherSystemMetrics(scope.NewSubScope("cloudevents_publisher")),
		events:        eventSet,
	}
}

func NewCloudEventsWrappedPublisher(
	db repositoryInterfaces.Repository, sender interfaces.Sender, scope promutils.Scope, storageClient *storage.DataStore, urlData dataInterfaces.RemoteURLInterface, remoteDataConfig runtimeInterfaces.RemoteDataConfig, eventPublisherConfig runtimeInterfaces.EventsPublisherConfig) interfaces.Publisher {

	return &CloudEventWrappedPublisher{
		db:                   db,
		sender:               sender,
		systemMetrics:        implementations.NewEventPublisherSystemMetrics(scope.NewSubScope("cloudevents_publisher")),
		storageClient:        storageClient,
		urlData:              urlData,
		remoteDataConfig:     remoteDataConfig,
		eventPublisherConfig: eventPublisherConfig,
	}
}
