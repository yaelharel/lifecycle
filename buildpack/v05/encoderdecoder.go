package v05

import (
	"os"

	"github.com/BurntSushi/toml"

	"github.com/buildpacks/lifecycle/api"
)

type Layer interface {
	Metadata() interface{}
	Build()    bool
	Launch()   bool
	Cache()    bool
}

type Layer05 struct {
	data interface{} `json:"data" toml:"metadata"`
	build  bool       `json:"build" toml:"build"`
	launch bool       `json:"launch" toml:"launch"`
	cache  bool       `json:"cache" toml:"cache"`
}

func (l *Layer05) Metadata() interface{}{
	return l.data
}

func (l *Layer05) Build() bool {
	return l.build
}

func (l *Layer05) Cache() bool {
	return l.cache
}

func (l *Layer05) Launch() bool {
	return l.launch
}

type EncoderDecoder05 struct {
}

func NewEncoderDecoder() *EncoderDecoder05 {
	return &EncoderDecoder05{}
}

func (d *EncoderDecoder05) IsSupported(buildpackAPI string) bool {
	return api.MustParse(buildpackAPI).Compare(api.MustParse("0.6")) < 0
}

func (d *EncoderDecoder05) Encode(file *os.File, lmf Layer) error {
	return toml.NewEncoder(file).Encode(lmf)
}

func (d *EncoderDecoder05) Decode(path string) (Layer, string, error) {
	var lmf Layer05
	md, err := toml.DecodeFile(path, &lmf)
	if err != nil {
		return &Layer05{}, "", err
	}
	msg := ""
	if isWrongFormat := typesInTypesTable(md); isWrongFormat {
		msg = "Warning: types table isn't supported in this buildpack api version. The launch, build and cache flags should be in the top level. Ignoring the values in the types table."
	}
	return &lmf, msg, nil
}

func typesInTypesTable(md toml.MetaData) bool {
	return md.IsDefined("types")
}
