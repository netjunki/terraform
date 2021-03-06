package remoteexec

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"reflect"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceProvider_Validate_good(t *testing.T) {
	c := testConfig(t, map[string]interface{}{
		"inline": "echo foo",
	})
	p := Provisioner()
	warn, errs := p.Validate(c)
	if len(warn) > 0 {
		t.Fatalf("Warnings: %v", warn)
	}
	if len(errs) > 0 {
		t.Fatalf("Errors: %v", errs)
	}
}

func TestResourceProvider_Validate_bad(t *testing.T) {
	c := testConfig(t, map[string]interface{}{
		"invalid": "nope",
	})
	p := Provisioner()
	warn, errs := p.Validate(c)
	if len(warn) > 0 {
		t.Fatalf("Warnings: %v", warn)
	}
	if len(errs) == 0 {
		t.Fatalf("Should have errors")
	}
}

var expectedScriptOut = `cd /tmp
wget http://foobar
exit 0
`

var expectedInlineScriptsOut = strings.Split(expectedScriptOut, "\n")

func TestResourceProvider_generateScript(t *testing.T) {
	p := Provisioner().(*schema.Provisioner)
	conf := map[string]interface{}{
		"inline": []interface{}{
			"cd /tmp",
			"wget http://foobar",
			"exit 0",
		},
	}
	out, err := generateScripts(schema.TestResourceDataRaw(
		t, p.Schema, conf))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if reflect.DeepEqual(out, expectedInlineScriptsOut) {
		t.Fatalf("bad: %v", out)
	}
}

func TestResourceProvider_CollectScripts_inline(t *testing.T) {
	p := Provisioner().(*schema.Provisioner)
	conf := map[string]interface{}{
		"inline": []interface{}{
			"cd /tmp",
			"wget http://foobar",
			"exit 0",
		},
	}

	scripts, err := collectScripts(schema.TestResourceDataRaw(
		t, p.Schema, conf))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if len(scripts) != 3 {
		t.Fatalf("bad: %v", scripts)
	}

	for i, script := range scripts {
		var out bytes.Buffer
		_, err = io.Copy(&out, script)
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if out.String() != expectedInlineScriptsOut[i] {
			t.Fatalf("bad: %v", out.String())
		}
	}
}

func TestResourceProvider_CollectScripts_script(t *testing.T) {
	p := Provisioner().(*schema.Provisioner)
	conf := map[string]interface{}{
		"script": "test-fixtures/script1.sh",
	}

	scripts, err := collectScripts(schema.TestResourceDataRaw(
		t, p.Schema, conf))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if len(scripts) != 1 {
		t.Fatalf("bad: %v", scripts)
	}

	var out bytes.Buffer
	_, err = io.Copy(&out, scripts[0])
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if out.String() != expectedScriptOut {
		t.Fatalf("bad: %v", out.String())
	}
}

func TestResourceProvider_CollectScripts_scripts(t *testing.T) {
	p := Provisioner().(*schema.Provisioner)
	conf := map[string]interface{}{
		"scripts": []interface{}{
			"test-fixtures/script1.sh",
			"test-fixtures/script1.sh",
			"test-fixtures/script1.sh",
		},
	}

	scripts, err := collectScripts(schema.TestResourceDataRaw(
		t, p.Schema, conf))
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if len(scripts) != 3 {
		t.Fatalf("bad: %v", scripts)
	}

	for idx := range scripts {
		var out bytes.Buffer
		_, err = io.Copy(&out, scripts[idx])
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		if out.String() != expectedScriptOut {
			t.Fatalf("bad: %v", out.String())
		}
	}
}

func testConfig(
	t *testing.T,
	c map[string]interface{}) *terraform.ResourceConfig {
	r, err := config.NewRawConfig(c)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	return terraform.NewResourceConfig(r)
}
