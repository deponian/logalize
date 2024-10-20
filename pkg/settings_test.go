package logalize

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/muesli/termenv"
)

// func TestSettingsInit(t *testing.T) {
// 	colorProfile = termenv.TrueColor

// 	lemmatizer, err := golem.New(en.New())
// 	if err != nil {
// 		t.Errorf("golem.New(en.New()) failed with this error: %s", err)
// 	}

// 	Opts = Settings{
// 		Theme: "tokyonight",
// 	}

// 	err = InitConfig(builtinsAllGood)
// 	if err != nil {
// 		t.Errorf("InitConfig() failed with this error: %s", err)
// 	}

// 	for _, tt := range tests {
// 		testname := tt.plain
// 		input := strings.NewReader(tt.plain)
// 		output := bytes.Buffer{}

// 		t.Run(testname, func(t *testing.T) {
// 			err := Run(input, &output, lemmatizer)
// 			if err != nil {
// 				t.Errorf("Run() failed with this error: %s", err)
// 			}

// 			result := strings.TrimSuffix(output.String(), "\n")

// 			if result != tt.colored {
// 				t.Errorf("got %v, want %v", result, tt.colored)
// 			}
// 		})
// 	}
// }

func TestSettingsFromConfig(t *testing.T) {
	colorProfile = termenv.TrueColor

	configData := `
settings:
  theme: "test"

  no-builtin-logformats: true
  no-builtin-patterns: true
  no-builtin-words: true
  no-builtins: true

  only-logformats: true
  only-patterns: true
  only-words: true
`
	configRaw := []byte(configData)
	config := koanf.New(".")
	if err := config.Load(rawbytes.Provider(configRaw), yaml.Parser()); err != nil {
		t.Errorf("Error during config loading: %s", err)
	}

	correctOpts := Settings{
		Theme: "test",

		NoBuiltinLogFormats: true,
		NoBuiltinPatterns:   true,
		NoBuiltinWords:      true,
		NoBuiltins:          true,

		HighlightOnlyLogFormats: true,
		HighlightOnlyPatterns:   true,
		HighlightOnlyWords:      true,
	}

	t.Run("TestSettingsFromConfig", func(t *testing.T) {
		opts := getSettingFromConfig(Settings{}, config)

		if !cmp.Equal(opts, correctOpts) {
			t.Errorf("got %v, want %v", opts, correctOpts)
		}
	})
}
