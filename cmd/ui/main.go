package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/alibaba/morphling/pkg/ui"
)

var (
	port, host, buildDir *string
)

func init() {
	port = flag.String("port", "80", "the port to listen to for incoming HTTP connections")
	host = flag.String("host", "0.0.0.0", "the host to listen to for incoming HTTP connections")
	buildDir = flag.String("build-dir", "/app/build", "the dir of frontend")
}
func main() {
	flag.Parse()
	kuh := ui.NewMorphlingUIHandler()

	log.Printf("Serving the frontend dir %s", *buildDir)
	frontend := http.FileServer(http.Dir(*buildDir))
	http.Handle("/morphling/", http.StripPrefix("/morphling/", frontend))

	http.HandleFunc("/morphling/fetch_hp_jobs/", kuh.FetchAllHPJobs)
	http.HandleFunc("/morphling/submit_profiling_yaml/", kuh.SubmitProfilingYamlJob)
	http.HandleFunc("/morphling/submit_trial_yaml/", kuh.SubmitTrialYamlJob)

	http.HandleFunc("/morphling/delete_experiment/", kuh.DeleteExperiment)
	http.HandleFunc("/morphling/fetch_experiment/", kuh.FetchExperiment)
	http.HandleFunc("/morphling/fetch_suggestion/", kuh.FetchSuggestion)

	http.HandleFunc("/morphling/fetch_hp_job_info/", kuh.FetchHPJobInfo)
	http.HandleFunc("/morphling/fetch_hp_job_trial_info/", kuh.FetchHPJobTrialInfo)
	//http.HandleFunc("/morphling/fetch_nas_job_info/", kuh.FetchNASJobInfo)

	http.HandleFunc("/morphling/submit_hp_job/", kuh.SubmitProfilingParametersJob)

	//http.HandleFunc("/morphling/fetch_trial_templates/", kuh.FetchTrialTemplates)
	//http.HandleFunc("/morphling/add_template/", kuh.AddTemplate)
	//http.HandleFunc("/morphling/edit_template/", kuh.EditTemplate)
	//http.HandleFunc("/morphling/delete_template/", kuh.DeleteTemplate)
	http.HandleFunc("/morphling/fetch_namespaces", kuh.FetchNamespaces)

	log.Printf("Serving at %s:%s", *host, *port)
	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", *host, *port), nil); err != nil {
		panic(err)
	}
}
