package utils

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/pkg/errors"
)

type ConfigParser struct {
	RawCfg     []byte
	validate   *validator.Validate
	translator ut.Translator
}

func NewConfigParser(configFile string) (*ConfigParser, error) {
	rawCfg, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config")
	}
	validate := validator.New()
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	err = enTranslations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		return nil, err
	}
	return &ConfigParser{RawCfg: rawCfg, validate: validate, translator: trans}, nil
}

func (c *ConfigParser) Parse(cfg any) error {
	err := yaml.Unmarshal(c.RawCfg, cfg)
	if err != nil {
		return errors.Wrap(err, "cannot parse config")
	}
	if err := c.validate.Struct(cfg); err != nil {
		return errors.New("config is invalid: " + c.translateErrors(err))
	}
	return nil
}

func (c *ConfigParser) translateErrors(err error) string {
	errs := []string{}
	for _, e := range err.(validator.ValidationErrors) {
		errs = append(errs, e.Translate(c.translator))
	}
	return strings.Join(errs, ", ")
}
