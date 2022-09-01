package phoneaudio

import (
	"os"
	"text/template"

	"github.com/gordonklaus/portaudio"
)

func (pa *PhoneAudio) Test() {
	var tmpl = template.Must(template.New("").Parse(
		`{{. | len}} host APIs: {{range .}}
{{.Name}}
	{{if .DefaultInputDevice}}Default input device:   {{.DefaultInputDevice.Name}}{{end}}
	{{if .DefaultOutputDevice}}Default output device:  {{.DefaultOutputDevice.Name}}{{end}}
	Devices: {{range .Devices}}
		{{.Name}}       IN:{{.MaxInputChannels}} OUT: {{.MaxOutputChannels}} Sample Rate: {{.DefaultSampleRate}}{{end}}
{{end}}`,
	))

	hs, err := portaudio.HostApis()
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(os.Stdout, hs); err != nil {
		panic(err)
	}
}
