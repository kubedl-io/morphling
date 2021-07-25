package grid

import (
	"fmt"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"testing"
	"time"

	suggestionapi "github.com/alibaba/morphling/api/v1alpha1/grpc/go"
	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func init() {
	logf.SetLogger(zap.New())

}

var (
	log     = logf.Log.WithName("sampling-client")
	timeout = 60 * time.Second
)

func TestValidateAlgorithmSettings(t *testing.T) {

	mega := gomega.NewGomegaWithT(t)
	// Get the service addr and dial it
	endpoint := "localhost:9996"
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	mega.Expect(err).NotTo(gomega.HaveOccurred())
	defer conn.Close()

	client := suggestionapi.NewSuggestionClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	request := newValidationRequest()
	msg, err := client.ValidateAlgorithmSettings(ctx, request, grpc.WaitForReady(true))
	mega.Expect(err).NotTo(gomega.HaveOccurred())
	log.Info(msg.String())

	statusCode, _ := status.FromError(err)

	// validation error
	if statusCode.Code() == codes.InvalidArgument || statusCode.Code() == codes.Unknown {
		log.Error(err, "ValidateAlgorithmSettings error")
	}

	// Connection error
	if statusCode.Code() == codes.Unavailable {
		log.Error(err, "Connection to Suggestion algorithm service currently unavailable")
	}

	// Validate to true as function is not implemented
	if statusCode.Code() == codes.Unimplemented {
		log.Info("Method ValidateAlgorithmSettings not found", "Suggestion service", request.AlgorithmName)
	}
	log.Info("Algorithm settings validated")
}

func TestSampling(t *testing.T) {

	mega := gomega.NewGomegaWithT(t)
	// Get the service addr and dial it
	endpoint := "localhost:9996"
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	mega.Expect(err).NotTo(gomega.HaveOccurred())
	defer conn.Close()

	client := suggestionapi.NewSuggestionClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	request := newSamplingRequest()
	response, err := client.GetSuggestions(ctx, request, grpc.WaitForReady(true))
	mega.Expect(err).NotTo(gomega.HaveOccurred())
	log.Info(response.String())

	statusCode, _ := status.FromError(err)

	// validation error
	if statusCode.Code() == codes.InvalidArgument || statusCode.Code() == codes.Unknown {
		log.Error(err, "ValidateAlgorithmSettings error")
	}

	// Connection error
	if statusCode.Code() == codes.Unavailable {
		log.Error(err, "Connection to Suggestion algorithm service currently unavailable")
	}

	// Validate to true as function is not implemented
	if statusCode.Code() == codes.Unimplemented {
		log.Info("Method ValidateAlgorithmSettings not found", "Suggestion service", request.AlgorithmName)
	}
	log.Info("Algorithm settings validated")

	Assignment := make([]morphlingv1alpha1.TrialAssignment, 0)
	for _, t := range response.AssignmentsSet {
		Assignment = append(Assignment,
			morphlingv1alpha1.TrialAssignment{
				Name:                 fmt.Sprintf("%s-%s", "test-sampling", utilrand.String(8)), // grid id
				ParameterAssignments: composeParameterAssignments(t.KeyValues),
			})
	}
}

func newValidationRequest() *suggestionapi.SamplingValidationRequest {
	request := &suggestionapi.SamplingValidationRequest{
		AlgorithmName:           "random",
		AlgorithmExtraSettings:  nil,
		SamplingNumberSpecified: 3,
		IsMaximize:              false,
	}
	pars := make([]suggestionapi.ParameterSpec, 0)
	pars = append(pars, suggestionapi.ParameterSpec{
		Name:          "cpu",
		ParameterType: suggestionapi.ParameterType_INT,
		FeasibleSpace: []string{"1", "2", "3.5"},
	})
	pars = append(pars, suggestionapi.ParameterSpec{
		Name:          "memory",
		ParameterType: suggestionapi.ParameterType_CATEGORICAL,
		FeasibleSpace: []string{"10G", "20G"},
	})
	return request
}

func newSamplingRequest() *suggestionapi.SamplingRequest {
	request := &suggestionapi.SamplingRequest{
		IsFirstRequest:          true,
		AlgorithmName:           "grid",
		AlgorithmExtraSettings:  nil,
		SamplingNumberSpecified: 10,
		RequiredSampling:        3,
		IsMaximize:              false,
		//ExistingResults:         nil,
	}
	pars := make([]*suggestionapi.ParameterSpec, 0)
	pars = append(pars, &suggestionapi.ParameterSpec{
		Name:          "cpu",
		ParameterType: suggestionapi.ParameterType_INT,
		FeasibleSpace: []string{"1", "2", "3.5", "0.5"},
	})
	pars = append(pars, &suggestionapi.ParameterSpec{
		Name:          "memory",
		ParameterType: suggestionapi.ParameterType_CATEGORICAL,
		FeasibleSpace: []string{"10G", "20G", "05G"},
	})
	request.Parameters = pars
	
	existingTrials := make([]*suggestionapi.TrialResult, 0)
	trial1 := &suggestionapi.TrialResult{
		ParameterAssignments: []*suggestionapi.KeyValue{{
			Key:   "cpu",
			Value: "0.5",
		},
		{
			Key:   "memory",
			Value: "05G",
		}},
		ObjectValue:          0,
	}
	existingTrials = append(existingTrials, trial1)
	request.ExistingResults = existingTrials
	return request
}

func composeParameterAssignments(pas []*suggestionapi.KeyValue) []morphlingv1alpha1.ParameterAssignment {
	res := make([]morphlingv1alpha1.ParameterAssignment, 0)
	for _, pa := range pas {
		categoryThis := morphlingv1alpha1.CategoryResource

		res = append(res, morphlingv1alpha1.ParameterAssignment{
			Name:     pa.Key,
			Value:    pa.Value,
			Category: categoryThis,
			//todo: Category
		})
	}
	return res
}
