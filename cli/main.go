package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/naoina/toml"
	"github.com/spf13/viper"

	"github.com/concrete-eth/archetype/codegen"
	"github.com/concrete-eth/archetype/codegen/gogen"
	"github.com/concrete-eth/archetype/codegen/solgen"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	GO_BIN              = "go"
	CONCRETE_BIN        = "concrete"
	GOFMT_BIN           = "gofmt"
	NODE_EXEC_BIN       = "npx"
	PRETTIER_BIN        = "prettier"
	PRETTIER_SOL_PLUGIN = "prettier-plugin-solidity"
	FORGE_BIN           = "forge"
	ABIGEN_BIN          = "abigen"
)

var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	gray   = color.New(color.FgHiBlack)
	yellow = color.New(color.FgYellow)
	bold   = color.New(color.Bold)
)

/* Logging */

func logTaskSuccess(name string, more ...any) {
	green.Print("[DONE] ")
	fmt.Print(name)
	if len(more) > 0 {
		gray.Print(": ")
		gray.Print(more...)
	}
	fmt.Println()
}

func logTaskFail(name string, err error) {
	red.Print("[FAIL] ")
	fmt.Print(name)
	if err != nil {
		fmt.Print(": ", err)
	}
	fmt.Println()
}

func logInfo(a ...any) {
	fmt.Println(a...)
}

func logDebug(a ...any) {
	gray.Println(a...)
}

func logWarning(warning string) {
	yellow.Println("\nWarning:")
	fmt.Println(warning)
}

func logError(err error) {
	fmt.Println("\nError:")
	red.Println(err)
	fmt.Println("\nContext:")
	logDebug(string(debug.Stack()))
	os.Exit(1)
}

func logFatal(err error) {
	logError(err)
	os.Exit(1)
}

/* Environment utils */

// ensureDir creates a directory if it does not exist.
func ensureDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory: %s", dir)
	}
	return nil
}

// isInstalled checks if a command is installed by attempting to run it with a help flag (-h, --help, help).
func isInstalled(name string, args ...string) bool {
	// Attempt to run the command with a help flag
	for _, flag := range []string{"-h", "--help", "help"} {
		args := append(args, flag)
		if err := exec.Command(name, args...).Run(); err == nil {
			// If the command runs without error, it is installed
			return true
		}
	}
	return false
}

// isInGoModule checks if the current directory is in a go module.
func isInGoModule() bool {
	cmd := exec.Command(GO_BIN, "env", "GOMOD")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	gomod := strings.TrimSpace(out.String())
	return gomod != "" && gomod != "/dev/null"
}

// getGoModule returns the name of the go module in the current directory.
func getGoModule() (string, error) {
	if !isInGoModule() {
		return "", fmt.Errorf("not in a go module")
	}
	cmd := exec.Command(GO_BIN, "list", "-m")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// getGoModulePath returns the root path of the go module in the current directory.
func getGoModulePath() (string, error) {
	if !isInGoModule() {
		return "", fmt.Errorf("not in a go module")
	}
	cmd := exec.Command(GO_BIN, "list", "-m", "-f", "{{.Dir}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

/* Verbose */

// loadSchemaFromFile loads a table schema from a json file.
func loadSchemaFromFile(filePath string) ([]datamod.TableSchema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return datamod.UnmarshalTableSchemas(data, false)
}

// printSchemaDescription prints a description of a table schema.
func printSchemaDescription(title string, schema []datamod.TableSchema) {
	description := codegen.GenerateSchemaDescriptionString(schema)
	bold.Println(title)
	fmt.Println(description)
}

/* Codegen */

// getGogenConfig returns a gogen config from viper settings.
func getGogenConfig() (gogen.Config, error) {
	var (
		actions = viper.GetString("actions")
		tables  = viper.GetString("tables")
		goOut   = viper.GetString("go-out")
		pkg     = viper.GetString("pkg")
		exp     = viper.GetBool("more-experimental")
	)

	var err error

	gogenOut := filepath.Join(goOut, "archmod") // <go-out>/archmod
	datamodOut := getDatamodOut()               // <go-out>/datamod
	datamodOutAbs, err := filepath.Abs(datamodOut)
	if err != nil {
		return gogen.Config{}, err
	}

	// Craft the import path for datamod e.g. github.com/user/repo/codegen/datamod
	var modName, modPath, relDatamodPath string
	if modName, err = getGoModule(); err != nil {
		return gogen.Config{}, err
	}
	if modPath, err = getGoModulePath(); err != nil {
		return gogen.Config{}, err
	}
	if relDatamodPath, err = filepath.Rel(modPath, datamodOutAbs); err != nil {
		return gogen.Config{}, err
	}
	datamodPkg := filepath.Join(modName, relDatamodPath)

	config := gogen.Config{
		Config: codegen.Config{
			Actions: actions,
			Tables:  tables,
			Out:     gogenOut,
		},
		Package:      pkg,
		Datamod:      datamodPkg,
		Experimental: exp,
	}

	return config, nil
}

// getDatamodOut returns the output directory for the datamod package.
func getDatamodOut() string {
	goOut := viper.GetString("go-out")
	datamodOut := filepath.Join(goOut, "datamod") // <go-out>/datamod
	return datamodOut
}

func getForgeBuildOut() string {
	forgeOut := viper.GetString("forge-out")
	return forgeOut
}

func getAbigenOut() string {
	goOut := viper.GetString("go-out")
	abigenOut := filepath.Join(goOut, "abigen") // <go-out>/abigen
	return abigenOut
}

// getSolgenConfig returns a solgen config from viper settings.
func getSolgenConfig() solgen.Config {
	var (
		actions = viper.GetString("actions")
		tables  = viper.GetString("tables")
		solOut  = viper.GetString("sol-out")
	)
	config := solgen.Config{
		Config: codegen.Config{
			Actions: actions,
			Tables:  tables,
			Out:     solOut,
		},
	}
	return config
}

// runCodegen runs the full code generation process.
func runCodegen(cmd *cobra.Command, args []string) {
	startTime := time.Now()
	verbose := viper.GetBool("verbose")

	if verbose {
		// Print settings
		allSettings := viper.AllSettings()
		settingsToml, err := toml.Marshal(allSettings)
		if err != nil {
			logFatal(err)
		}
		logDebug(string(settingsToml))
	}

	// Get and validate codegen configs
	// Go codegen config
	gogenConfig, err := getGogenConfig()
	if err != nil {
		logFatal(err)
	}
	if err := ensureDir(gogenConfig.Out); err != nil {
		logFatal(err)
	}
	if err := gogenConfig.Validate(); err != nil {
		logFatal(err)
	}
	// Solidity codegen config
	solgenConfig := getSolgenConfig()
	if err := ensureDir(solgenConfig.Out); err != nil {
		logFatal(err)
	}
	if err := solgenConfig.Validate(); err != nil {
		logFatal(err)
	}

	if verbose {
		// Print schema descriptions
		actionsSchema, err := loadSchemaFromFile(gogenConfig.Actions)
		if err != nil {
			logFatal(err)
		}
		tablesSchema, err := loadSchemaFromFile(gogenConfig.Tables)
		if err != nil {
			logFatal(err)
		}
		printSchemaDescription("Actions", actionsSchema)
		fmt.Println("")
		printSchemaDescription("Tables", tablesSchema)
		fmt.Println("")
	}

	// Preliminary checks
	if !isInstalled(GO_BIN) {
		logFatal(fmt.Errorf("go is not installed (go_bin=%s)", GO_BIN))
	}
	if !isInGoModule() {
		logFatal(fmt.Errorf("not in a go module"))
	}
	if !isInstalled(CONCRETE_BIN) {
		logFatal(fmt.Errorf("concrete cli is not installed (concrete_bin=%s)", CONCRETE_BIN))
	}
	if !isInstalled(FORGE_BIN) {
		logFatal(fmt.Errorf("forge cli is not installed (forge_bin=%s)", FORGE_BIN))
	}
	if !isInstalled(ABIGEN_BIN) {
		logFatal(fmt.Errorf("abigen is not installed (abigen_bin=%s)", ABIGEN_BIN))
	}

	// Run concrete datamod
	datamodPkg := "datamod"
	datamodOut := getDatamodOut()
	if err := ensureDir(datamodOut); err != nil {
		logFatal(err)
	}
	if err := runDatamod(datamodOut, gogenConfig.Tables, datamodPkg, gogenConfig.Experimental); err != nil {
		logFatal(err)
	}
	// Run go and solidity codegen
	if err := runGogen(gogenConfig); err != nil {
		logFatal(err)
	}
	if err := runSolgen(solgenConfig); err != nil {
		logFatal(err)
	}

	// Run gofmt
	if isInstalled(GOFMT_BIN) {
		runGofmt(datamodOut, gogenConfig.Out)
	} else {
		logWarning(fmt.Sprintf("gofmt is not installed (gofmt_bin=%s). Install it to format the generated go code.", GOFMT_BIN))
	}

	var missing string
	if isInstalled(NODE_EXEC_BIN) {
		if isInstalled(NODE_EXEC_BIN, PRETTIER_BIN) {
			if isInstalled(NODE_EXEC_BIN, PRETTIER_BIN, "--plugin="+PRETTIER_SOL_PLUGIN) {
				runPrettier(solgenConfig.Out + "/**/*.sol")
			} else {
				missing = PRETTIER_SOL_PLUGIN
			}
		} else {
			missing = PRETTIER_BIN
		}
	} else {
		missing = NODE_EXEC_BIN
	}
	if missing != "" {
		logWarning(fmt.Sprintf("%s is not installed. Install it to format the generated solidity code.", missing))
	}

	// Forge build
	forgeBuildOut := getForgeBuildOut()
	if err := runForgeBuild(solgenConfig.Out, forgeBuildOut); err != nil {
		logFatal(err)
	}

	// Abigen
	abigenOut := getAbigenOut()
	for _, contractName := range []string{"IActions", "ITables"} {
		var (
			inPath   = filepath.Join(forgeBuildOut, contractName+".sol")
			dirName  = strings.ToLower(contractName)
			fileName = dirName + ".go"
			outDir   = filepath.Join(abigenOut, dirName)
			outPath  = filepath.Join(outDir, fileName)
		)
		if err := ensureDir(outDir); err != nil {
			logFatal(err)
		}
		if err := runAbigen(contractName, inPath, outPath); err != nil {
			logFatal(err)
		}
	}

	// Done
	logInfo("\nCode generation completed successfully.")
	logInfo("Files written to: " + gogenConfig.Out + ", " + solgenConfig.Out)
	logDebug(fmt.Sprintf("\nDone in %v", time.Since(startTime)))
}

func runCommand(name string, cmd *exec.Cmd) error {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("%s failed: %w", name, err)
		logTaskFail(name, err)
		logDebug(strings.Join(cmd.Args, " "))
		logDebug(stdErr.String())
		return err
	}
	if viper.GetBool("verbose") {
		logTaskSuccess(name, strings.Join(cmd.Args, " "))
	} else {
		logTaskSuccess(name)
	}
	return nil
}

// Run concrete datamod.
// Datamod generates type safe go wrappers for datastore structures from a JSON specification.
func runDatamod(outDir, tables, pkg string, experimental bool) error {
	taskName := "Concrete datamod"
	args := []string{"datamod", tables, "--pkg", pkg, "--out", outDir}
	if experimental {
		args = append(args, "--more-experimental")
	}
	cmd := exec.Command(CONCRETE_BIN, args...)
	return runCommand(taskName, cmd)
}

// Run gogen codegen
// config is assumed to be valid.
func runGogen(config gogen.Config) error {
	taskName := "Go"
	if err := gogen.Codegen(config); err != nil {
		err = fmt.Errorf("gogen failed: %w", err)
		logTaskFail(taskName, nil)
		return err
	}
	logTaskSuccess(taskName)
	return nil
}

// Run solgen codegen
// config is assumed to be valid
func runSolgen(config solgen.Config) error {
	taskName := "Solidity"
	if err := solgen.Codegen(config); err != nil {
		err = fmt.Errorf("solgen failed: %w", err)
		logTaskFail(taskName, nil)
		return err
	}
	logTaskSuccess(taskName)
	return nil
}

func runForgeBuild(inDir, outDir string) error {
	taskName := "Forge build"
	cmd := exec.Command(
		FORGE_BIN, "build",
		"--contracts", inDir,
		"--out", outDir,
		"--extra-output-files", "bin", "abi",
	)
	return runCommand(taskName, cmd)
}

func runAbigen(contractName, inPath, outPath string) error {
	var (
		pgk     = "contract"
		binPath = filepath.Join(inPath, contractName+".bin")
		abiDir  = filepath.Join(inPath, contractName+".abi.json")
	)
	taskName := "abigen: " + contractName
	cmd := exec.Command(ABIGEN_BIN, "--bin", binPath, "--abi", abiDir, "--pkg", pgk, "--out", outPath)
	return runCommand(taskName, cmd)
}

// runGofmt runs gofmt on the given directory.
func runGofmt(dirs ...string) error {
	taskName := "gofmt"
	args := append([]string{"-w"}, dirs...)
	cmd := exec.Command(GOFMT_BIN, args...)
	return runCommand(taskName, cmd)
}

// runPrettier runs prettier on the given directory.
func runPrettier(patterns ...string) error {
	taskName := "prettier"
	args := []string{PRETTIER_BIN, "--plugin=" + PRETTIER_SOL_PLUGIN, "--write"}
	args = append(args, patterns...)
	cmd := exec.Command(NODE_EXEC_BIN, args...)
	return runCommand(taskName, cmd)
}

/* CLI */

// NewRootCmd creates the root command for the CLI.
func NewRootCmd() *cobra.Command {
	// Root command
	var cfgFile string
	var rootCmd = &cobra.Command{
		Use: "archetype",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig(cfgFile)
		},
	}
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./arch.toml)")

	// Codegen command
	codegenCmd := &cobra.Command{Use: "codegen", Short: "Generate Golang definitions and Solidity interfaces for Archetype tables and actions from the given JSON specifications", Run: runCodegen}

	// Codegen flags
	codegenCmd.Flags().StringP("go-out", "g", "./codegen", "output directory")
	codegenCmd.Flags().StringP("sol-out", "s", "./codegen/sol", "output directory")
	codegenCmd.Flags().StringP("forge-out", "f", "./out", "forge output directory")
	codegenCmd.Flags().StringP("tables", "t", "./tables.json", "table schema file")
	codegenCmd.Flags().StringP("actions", "a", "./actions.json", "action schema file")
	codegenCmd.Flags().String("pkg", "archmod", "go package name")
	codegenCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	codegenCmd.Flags().Bool("more-experimental", false, "enable experimental features")

	// Bind flags to viper
	viper.BindPFlag("go-out", codegenCmd.Flags().Lookup("go-out"))
	viper.BindPFlag("sol-out", codegenCmd.Flags().Lookup("sol-out"))
	viper.BindPFlag("forge-out", codegenCmd.Flags().Lookup("forge-out"))
	viper.BindPFlag("tables", codegenCmd.Flags().Lookup("tables"))
	viper.BindPFlag("actions", codegenCmd.Flags().Lookup("actions"))
	viper.BindPFlag("pkg", codegenCmd.Flags().Lookup("pkg"))
	viper.BindPFlag("verbose", codegenCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("more-experimental", codegenCmd.Flags().Lookup("more-experimental"))

	rootCmd.AddCommand(codegenCmd)

	return rootCmd
}

// initConfig loads the viper configuration from the given or default file and the environment.
// See https://github.com/spf13/viper?tab=readme-ov-file#why-viper for precedence order.
func initConfig(cfgFile string) {
	// Get config from file
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in the working directory
		viper.AddConfigPath(".")
		viper.SetConfigName("arch")
	}

	// Get config from environment
	viper.SetEnvPrefix("ARCH")
	viper.AutomaticEnv()

	// Read config
	if err := viper.ReadInConfig(); err == nil {
		// Log the config file used
		configFileAbsPath := viper.ConfigFileUsed()
		wd, err := os.Getwd()
		if err != nil {
			logFatal(err)
		}
		configFileRelPath, err := filepath.Rel(wd, configFileAbsPath)
		if err != nil {
			logFatal(err)
		}
		var configFilePathToPrint string
		if strings.HasPrefix(configFileRelPath, "..") {
			// If the config file is outside the working directory, print the absolute path
			configFilePathToPrint = configFileAbsPath
		} else {
			// Otherwise, print the relative path
			configFilePathToPrint = "./" + configFileRelPath
		}
		logDebug("Using config file:", configFilePathToPrint)
		fmt.Println("")
	} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		logError(err)
		os.Exit(1)
	}
}

// Execute runs the CLI.
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		logFatal(err)
	}
}
