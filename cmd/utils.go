package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/asdine/storm/v3"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v3"

	settings "github.com/nulnl/nulyun/settings/global"
	"github.com/nulnl/nulyun/settings/users"
	"github.com/nulnl/nulyun/storage"
	"github.com/nulnl/nulyun/storage/bolt"
)

const databasePermissions = 0640

func getAndParseFileMode(flags *pflag.FlagSet, name string) (fs.FileMode, error) {
	mode, err := flags.GetString(name)
	if err != nil {
		return 0, err
	}

	b, err := strconv.ParseUint(mode, 0, 32)
	if err != nil {
		return 0, err
	}

	return fs.FileMode(b), nil
}

func generateKey() []byte {
	k, err := settings.GenerateKey()
	if err != nil {
		panic(err)
	}
	return k
}

func dbExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.Size() != 0, nil
	}

	if os.IsNotExist(err) {
		d := filepath.Dir(path)
		_, err = os.Stat(d)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(d, 0700); err != nil {
				return false, err
			}
			return false, nil
		}
	}

	return false, err
}

// Generate the replacements for all environment variables. This allows to
// use FB_BRANDING_DISABLE_EXTERNAL environment variables, even when the
// option name is branding.disableExternal.
func generateEnvKeyReplacements(cmd *cobra.Command) []string {
	replacements := []string{}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		oldName := strings.ToUpper(f.Name)
		newName := strings.ToUpper(lo.SnakeCase(f.Name))
		replacements = append(replacements, oldName, newName)
	})

	return replacements
}

func initViper(cmd *cobra.Command) (*viper.Viper, error) {
	v := viper.New()

	// Get config file from flag
	cfgFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}

	// Configuration file
	if cfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		v.AddConfigPath(".")
		v.AddConfigPath(home)
		v.AddConfigPath("/etc/nulyun/")
		v.SetConfigName(".nulyun")
	} else {
		v.SetConfigFile(cfgFile)
	}

	// Environment variables
	v.SetEnvPrefix("FB")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(generateEnvKeyReplacements(cmd)...))

	// Bind the flags
	err = v.BindPFlags(cmd.Flags())
	if err != nil {
		return nil, err
	}

	// Read in configuration
	if err := v.ReadInConfig(); err != nil {
		if errors.Is(err, viper.ConfigParseError{}) {
			return nil, err
		}

		log.Println("No config file used")
	} else {
		log.Printf("Using config file: %s", v.ConfigFileUsed())
	}

	// Return Viper
	return v, nil
}

type store struct {
	*storage.Storage
	databaseExisted bool
}

type storeOptions struct {
	expectsNoDatabase bool
	allowsNoDatabase  bool
}

type cobraFunc func(cmd *cobra.Command, args []string) error

// withViperAndStore initializes Viper and the storage.Store and passes them to the callback function.
// This function should only be used by [withStore] and the root command. No other command should call
// this function directly.
func withViperAndStore(fn func(cmd *cobra.Command, args []string, v *viper.Viper, store *store) error, options storeOptions) cobraFunc {
	return func(cmd *cobra.Command, args []string) error {
		v, err := initViper(cmd)
		if err != nil {
			return err
		}

		path, err := filepath.Abs(v.GetString("database"))
		if err != nil {
			return err
		}

		exists, err := dbExists(path)
		switch {
		case err != nil:
			return err
		case exists && options.expectsNoDatabase:
			log.Fatal(path + " already exists")
		case !exists && !options.expectsNoDatabase && !options.allowsNoDatabase:
			log.Fatal(path + " does not exist. Please run 'nulyun config init' first.")
		case !exists && !options.expectsNoDatabase:
			log.Println("WARNING: nulyun.db can't be found. Initialing in " + strings.TrimSuffix(path, "nulyun.db"))
		}

		log.Println("Using database: " + path)

		db, err := storm.Open(path, storm.BoltOptions(databasePermissions, nil))
		if err != nil {
			return err
		}
		defer db.Close()

		storage, err := bolt.NewStorage(db)
		if err != nil {
			return err
		}

		store := &store{
			Storage:         storage,
			databaseExisted: exists,
		}

		return fn(cmd, args, v, store)
	}
}

func withStore(fn func(cmd *cobra.Command, args []string, store *store) error, options storeOptions) cobraFunc {
	return withViperAndStore(func(cmd *cobra.Command, args []string, _ *viper.Viper, store *store) error {
		return fn(cmd, args, store)
	}, options)
}

func marshal(filename string, data interface{}) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	switch ext := filepath.Ext(filename); ext {
	case ".json":
		encoder := json.NewEncoder(fd)
		encoder.SetIndent("", "    ")
		return encoder.Encode(data)
	case ".yml", ".yaml":
		encoder := yaml.NewEncoder(fd)
		return encoder.Encode(data)
	default:
		return errors.New("invalid format: " + ext)
	}
}

func unmarshal(filename string, data interface{}) error {
	fd, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	switch ext := filepath.Ext(filename); ext {
	case ".json":
		return json.NewDecoder(fd).Decode(data)
	case ".yml", ".yaml":
		return yaml.NewDecoder(fd).Decode(data)
	default:
		return errors.New("invalid format: " + ext)
	}
}

func jsonYamlArg(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	switch ext := filepath.Ext(args[0]); ext {
	case ".json", ".yml", ".yaml":
		return nil
	default:
		return errors.New("invalid format: " + ext)
	}
}

func cleanUpInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		result[fmt.Sprintf("%v", k)] = cleanUpMapValue(v)
	}
	return result
}

func cleanUpInterfaceArray(in []interface{}) []interface{} {
	result := make([]interface{}, len(in))
	for i, v := range in {
		result[i] = cleanUpMapValue(v)
	}
	return result
}

func cleanUpMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanUpInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanUpInterfaceMap(v)
	default:
		return v
	}
}

// convertCmdStrToCmdArray checks if cmd string is blank (whitespace included)
// then returns empty string array, else returns the split word array of cmd.
// This is to ensure the result will never be []string{""}
func convertCmdStrToCmdArray(cmd string) []string {
	var cmdArray []string
	trimmedCmdStr := strings.TrimSpace(cmd)
	if trimmedCmdStr != "" {
		cmdArray = strings.Split(trimmedCmdStr, " ")
	}
	return cmdArray
}

// addUserFlags adds user-related flags used by config command.
func addUserFlags(flags *pflag.FlagSet) {
	flags.Bool("perm.admin", false, "admin perm for users")
	flags.Bool("perm.execute", true, "execute perm for users")
	flags.Bool("perm.create", true, "create perm for users")
	flags.Bool("perm.rename", true, "rename perm for users")
	flags.Bool("perm.modify", true, "modify perm for users")
	flags.Bool("perm.delete", true, "delete perm for users")
	flags.Bool("perm.share", true, "share perm for users")
	flags.Bool("perm.download", true, "download perm for users")
	flags.String("sorting.by", "name", "sorting mode (name, size or modified)")
	flags.Bool("sorting.asc", false, "sorting by ascending order")
	flags.Bool("lockPassword", false, "lock password")
	flags.StringSlice("commands", nil, "a list of the commands a user can execute")
	flags.String("scope", ".", "scope for users")
	flags.String("locale", "en", "locale for users")
	flags.String("viewMode", string(users.ListViewMode), "view mode for users")
	flags.Bool("singleClick", false, "use single clicks only")
	flags.Bool("dateFormat", false, "use date format (true for absolute time, false for relative)")
	flags.Bool("hideDotfiles", false, "hide dotfiles")
	flags.String("aceEditorTheme", "", "ace editor's syntax highlighting theme for users")
}

func getAndParseViewMode(flags *pflag.FlagSet) (users.ViewMode, error) {
	viewModeStr, err := flags.GetString("viewMode")
	if err != nil {
		return "", err
	}

	viewMode := users.ViewMode(viewModeStr)
	if viewMode != users.ListViewMode && viewMode != users.MosaicViewMode {
		return "", errors.New("view mode must be \"" + string(users.ListViewMode) + "\" or \"" + string(users.MosaicViewMode) + "\"")
	}

	return viewMode, nil
}

// getUserDefaults retrieves user default settings from flags.
func getUserDefaults(flags *pflag.FlagSet, defaults *settings.UserDefaults, all bool) error {
	errs := []error{}

	visit := func(flag *pflag.Flag) {
		var err error
		switch flag.Name {
		case "scope":
			defaults.Scope, err = flags.GetString(flag.Name)
		case "locale":
			defaults.Locale, err = flags.GetString(flag.Name)
		case "viewMode":
			defaults.ViewMode, err = getAndParseViewMode(flags)
		case "singleClick":
			defaults.SingleClick, err = flags.GetBool(flag.Name)
		case "aceEditorTheme":
			defaults.AceEditorTheme, err = flags.GetString(flag.Name)
		case "perm.admin":
			defaults.Perm.Admin, err = flags.GetBool(flag.Name)
		case "perm.execute":
			defaults.Perm.Execute, err = flags.GetBool(flag.Name)
		case "perm.create":
			defaults.Perm.Create, err = flags.GetBool(flag.Name)
		case "perm.rename":
			defaults.Perm.Rename, err = flags.GetBool(flag.Name)
		case "perm.modify":
			defaults.Perm.Modify, err = flags.GetBool(flag.Name)
		case "perm.delete":
			defaults.Perm.Delete, err = flags.GetBool(flag.Name)
		case "perm.share":
			defaults.Perm.Share, err = flags.GetBool(flag.Name)
		case "perm.download":
			defaults.Perm.Download, err = flags.GetBool(flag.Name)
		case "sorting.by":
			defaults.Sorting.By, err = flags.GetString(flag.Name)
		case "sorting.asc":
			defaults.Sorting.Asc, err = flags.GetBool(flag.Name)
		case "hideDotfiles":
			defaults.HideDotfiles, err = flags.GetBool(flag.Name)
		}

		if err != nil {
			errs = append(errs, err)
		}
	}

	if all {
		flags.VisitAll(visit)
	} else {
		flags.Visit(visit)
	}

	return errors.Join(errs...)
}
