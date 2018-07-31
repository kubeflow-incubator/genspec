package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kubeflow-incubator/genspec/pkg/spec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	outputPath string
)

var rootCmd = &cobra.Command{
	Use:   "genspec",
	Short: "Generate OpenAPI specification for TFJob",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		generateSwagger()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&outputPath, "output", "swagger.json", "Path to write OpenAPI spec file")
}

// cleanupConfig cleans up temporary config file `apiserver.local.config` which created during generating swagger.
func cleanupConfig() {
	_, err := os.Stat(spec.ApiserverLocalConfigPath)
	if err == nil {
		os.RemoveAll(spec.ApiserverLocalConfigPath)
	}
}

// generateSwagger generates OpenAPI swagger and write to file.
func generateSwagger() {
	defer cleanupConfig()

	apiSpec, err := spec.RenderSwaggerJson()
	if err != nil {
		log.Fatalf("Failed to generate spec: %v", err)
	}

	err = ioutil.WriteFile(outputPath, []byte(apiSpec), 0644)
	if err != nil {
		log.Fatalf("Failed to write spec: %v", err)
	} else {
		log.Infof("Write swagger to %v successful", outputPath)
	}
}
