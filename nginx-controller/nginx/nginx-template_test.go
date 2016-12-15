package nginx

import (
	"path/filepath"
	"html/template"
	"os"
	"testing"
	"bytes"
	"github.com/stretchr/testify/assert"
)


func TestNginxTemplate(t *testing.T) {
	assert := assert.New(t)

	cwd, _ := os.Getwd()
	templatePath := filepath.Join(cwd, "ingress.tmpl")

	tmpl, err := template.New("ingress.tmpl").ParseFiles(templatePath)

	if err != nil {
		println("Failed to parse template file")
	}


	var servers []Server
	var upstreams []Upstream

	upstream := Upstream{
		Name: "Upstream-Name",
		UpstreamServers: [] UpstreamServer{
			{Address:"Upstream-Address", Port: "8080"},
		},
	}


	upstreams = append(upstreams, upstream)

	server := Server{
		Name: "Test",
		Locations: []Location{
			{Path: "path",Upstream:upstream},
		},
	}

	servers = append(servers, server)

	config := IngressNginxConfig{
		Servers: servers,
		Upstreams: upstreams,
		Labels: map[string]string{
			"FAKESSL": "TRUE",
			"MyLabel": "TESTLAbelValue",
		},
	}

	var out bytes.Buffer
	tmpl.Execute(os.Stdout, config)
	tmpl.Execute(&out, config)
	var result = out.String()

	var expected = `
upstream Upstream-Name {

	server Upstream-Address:8080;
}


server {
	listen 80;


        if ($http_X_ISSSL != 'TRUE') {
                return 301 https://$host$request_uri;
        }




	server_name Test;




	location path {
		proxy_http_version 1.1;



		proxy_connect_timeout ;
		proxy_read_timeout ;
		client_max_body_size ;
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header X-Forwarded-Host $host;
		proxy_set_header X-Forwarded-Port $server_port;
		proxy_set_header X-Forwarded-Proto $scheme;
		proxy_pass http://Upstream-Name;
	}
}
	`

	assert.Equal(result, expected, "Should parse labels")
}
