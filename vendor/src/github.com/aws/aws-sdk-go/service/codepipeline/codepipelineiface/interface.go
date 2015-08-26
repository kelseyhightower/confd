// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

// Package codepipelineiface provides an interface for the CodePipeline Service.
package codepipelineiface

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/codepipeline"
)

// CodePipelineAPI is the interface type for codepipeline.CodePipeline.
type CodePipelineAPI interface {
	AcknowledgeJobRequest(*codepipeline.AcknowledgeJobInput) (*request.Request, *codepipeline.AcknowledgeJobOutput)

	AcknowledgeJob(*codepipeline.AcknowledgeJobInput) (*codepipeline.AcknowledgeJobOutput, error)

	AcknowledgeThirdPartyJobRequest(*codepipeline.AcknowledgeThirdPartyJobInput) (*request.Request, *codepipeline.AcknowledgeThirdPartyJobOutput)

	AcknowledgeThirdPartyJob(*codepipeline.AcknowledgeThirdPartyJobInput) (*codepipeline.AcknowledgeThirdPartyJobOutput, error)

	CreateCustomActionTypeRequest(*codepipeline.CreateCustomActionTypeInput) (*request.Request, *codepipeline.CreateCustomActionTypeOutput)

	CreateCustomActionType(*codepipeline.CreateCustomActionTypeInput) (*codepipeline.CreateCustomActionTypeOutput, error)

	CreatePipelineRequest(*codepipeline.CreatePipelineInput) (*request.Request, *codepipeline.CreatePipelineOutput)

	CreatePipeline(*codepipeline.CreatePipelineInput) (*codepipeline.CreatePipelineOutput, error)

	DeleteCustomActionTypeRequest(*codepipeline.DeleteCustomActionTypeInput) (*request.Request, *codepipeline.DeleteCustomActionTypeOutput)

	DeleteCustomActionType(*codepipeline.DeleteCustomActionTypeInput) (*codepipeline.DeleteCustomActionTypeOutput, error)

	DeletePipelineRequest(*codepipeline.DeletePipelineInput) (*request.Request, *codepipeline.DeletePipelineOutput)

	DeletePipeline(*codepipeline.DeletePipelineInput) (*codepipeline.DeletePipelineOutput, error)

	DisableStageTransitionRequest(*codepipeline.DisableStageTransitionInput) (*request.Request, *codepipeline.DisableStageTransitionOutput)

	DisableStageTransition(*codepipeline.DisableStageTransitionInput) (*codepipeline.DisableStageTransitionOutput, error)

	EnableStageTransitionRequest(*codepipeline.EnableStageTransitionInput) (*request.Request, *codepipeline.EnableStageTransitionOutput)

	EnableStageTransition(*codepipeline.EnableStageTransitionInput) (*codepipeline.EnableStageTransitionOutput, error)

	GetJobDetailsRequest(*codepipeline.GetJobDetailsInput) (*request.Request, *codepipeline.GetJobDetailsOutput)

	GetJobDetails(*codepipeline.GetJobDetailsInput) (*codepipeline.GetJobDetailsOutput, error)

	GetPipelineRequest(*codepipeline.GetPipelineInput) (*request.Request, *codepipeline.GetPipelineOutput)

	GetPipeline(*codepipeline.GetPipelineInput) (*codepipeline.GetPipelineOutput, error)

	GetPipelineStateRequest(*codepipeline.GetPipelineStateInput) (*request.Request, *codepipeline.GetPipelineStateOutput)

	GetPipelineState(*codepipeline.GetPipelineStateInput) (*codepipeline.GetPipelineStateOutput, error)

	GetThirdPartyJobDetailsRequest(*codepipeline.GetThirdPartyJobDetailsInput) (*request.Request, *codepipeline.GetThirdPartyJobDetailsOutput)

	GetThirdPartyJobDetails(*codepipeline.GetThirdPartyJobDetailsInput) (*codepipeline.GetThirdPartyJobDetailsOutput, error)

	ListActionTypesRequest(*codepipeline.ListActionTypesInput) (*request.Request, *codepipeline.ListActionTypesOutput)

	ListActionTypes(*codepipeline.ListActionTypesInput) (*codepipeline.ListActionTypesOutput, error)

	ListPipelinesRequest(*codepipeline.ListPipelinesInput) (*request.Request, *codepipeline.ListPipelinesOutput)

	ListPipelines(*codepipeline.ListPipelinesInput) (*codepipeline.ListPipelinesOutput, error)

	PollForJobsRequest(*codepipeline.PollForJobsInput) (*request.Request, *codepipeline.PollForJobsOutput)

	PollForJobs(*codepipeline.PollForJobsInput) (*codepipeline.PollForJobsOutput, error)

	PollForThirdPartyJobsRequest(*codepipeline.PollForThirdPartyJobsInput) (*request.Request, *codepipeline.PollForThirdPartyJobsOutput)

	PollForThirdPartyJobs(*codepipeline.PollForThirdPartyJobsInput) (*codepipeline.PollForThirdPartyJobsOutput, error)

	PutActionRevisionRequest(*codepipeline.PutActionRevisionInput) (*request.Request, *codepipeline.PutActionRevisionOutput)

	PutActionRevision(*codepipeline.PutActionRevisionInput) (*codepipeline.PutActionRevisionOutput, error)

	PutJobFailureResultRequest(*codepipeline.PutJobFailureResultInput) (*request.Request, *codepipeline.PutJobFailureResultOutput)

	PutJobFailureResult(*codepipeline.PutJobFailureResultInput) (*codepipeline.PutJobFailureResultOutput, error)

	PutJobSuccessResultRequest(*codepipeline.PutJobSuccessResultInput) (*request.Request, *codepipeline.PutJobSuccessResultOutput)

	PutJobSuccessResult(*codepipeline.PutJobSuccessResultInput) (*codepipeline.PutJobSuccessResultOutput, error)

	PutThirdPartyJobFailureResultRequest(*codepipeline.PutThirdPartyJobFailureResultInput) (*request.Request, *codepipeline.PutThirdPartyJobFailureResultOutput)

	PutThirdPartyJobFailureResult(*codepipeline.PutThirdPartyJobFailureResultInput) (*codepipeline.PutThirdPartyJobFailureResultOutput, error)

	PutThirdPartyJobSuccessResultRequest(*codepipeline.PutThirdPartyJobSuccessResultInput) (*request.Request, *codepipeline.PutThirdPartyJobSuccessResultOutput)

	PutThirdPartyJobSuccessResult(*codepipeline.PutThirdPartyJobSuccessResultInput) (*codepipeline.PutThirdPartyJobSuccessResultOutput, error)

	StartPipelineExecutionRequest(*codepipeline.StartPipelineExecutionInput) (*request.Request, *codepipeline.StartPipelineExecutionOutput)

	StartPipelineExecution(*codepipeline.StartPipelineExecutionInput) (*codepipeline.StartPipelineExecutionOutput, error)

	UpdatePipelineRequest(*codepipeline.UpdatePipelineInput) (*request.Request, *codepipeline.UpdatePipelineOutput)

	UpdatePipeline(*codepipeline.UpdatePipelineInput) (*codepipeline.UpdatePipelineOutput, error)
}
