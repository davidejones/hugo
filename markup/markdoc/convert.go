// Copyright 2025 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package markdoc converts content to HTML using a Markdoc external helper
// implemented as a local Node.js script in node_modules.
package markdoc

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/gohugoio/hugo/common/hexec"
	"github.com/gohugoio/hugo/htesting"
	"github.com/gohugoio/hugo/identity"

	"github.com/gohugoio/hugo/markup/converter"
	"github.com/gohugoio/hugo/markup/internal"
)

// Provider is the package entry point.
var Provider converter.ProviderProvider = provider{}

type provider struct{}

func (p provider) New(cfg converter.ProviderConfig) (converter.Provider, error) {
	return converter.NewProvider("markdoc", func(ctx converter.DocumentContext) (converter.Converter, error) {
		return &markdocConverter{
			ctx: ctx,
			cfg: cfg,
		}, nil
	}), nil
}

type markdocConverter struct {
	ctx converter.DocumentContext
	cfg converter.ProviderConfig
}

func (c *markdocConverter) Convert(ctx converter.RenderContext) (converter.ResultRender, error) {
	b, err := c.getMarkdocContent(ctx.Src, c.ctx)
	if err != nil {
		return nil, err
	}
	return converter.Bytes(b), nil
}

func (c *markdocConverter) Supports(feature identity.Identity) bool {
	return false
}

// getMarkdocContent calls a local Node.js helper script (mdoc.js) in node_modules
// to convert Markdoc content to HTML.
func (c *markdocConverter) getMarkdocContent(src []byte, ctx converter.DocumentContext) ([]byte, error) {
	logger := c.cfg.Logger

	nodeBinary := getNodeBinaryName()
	if nodeBinary == "" {
		logger.Println("node not found in $PATH: Please install Node.js.\n",
			"                 Leaving Markdoc content unrendered.")
		return src, nil
	}

	// Try to locate the Markdoc CLI script in local node_modules.
	// We accept a few common locations; the user requested a script named mdoc.js.
	candidates := []string{
		filepath.FromSlash("local/bin/js/cdocs-hugo.js"),
	}

	var scriptPath string
	for _, cand := range candidates {
		// Check relative to the current working directory.
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			scriptPath = cand
			break
		}
		// Also allow it to be on PATH in some environments.
		if p := hexec.LookPath(cand); p != "" {
			scriptPath = p
			break
		}
	}
	if scriptPath == "" {
		// Fall back to node_modules/.bin/mdoc(.cmd) if present in PATH (e.g., via npm scripts environment)
		// Note: We cannot pass paths with slashes to internal.ExternallyRenderContent's binaryName,
		// so we always use the node binary and pass the script as the first argument.
		// If no script file can be found, leave content as-is.
		logger.Println("cdocs-hugo.js not found locally: Please install the Customizable Docs.\n",
			"                 Leaving Markdoc content unrendered.")
		return src, nil
	}

	logger.Infoln("Rendering", ctx.DocumentName, "with", "node", scriptPath, "...")

	args := []string{scriptPath}

	// Some CLIs behave differently on Windows; we keep it consistent by always using node to execute the JS file.
	_ = runtime.GOOS // currently unused, but left for parity with rst implementation comments

	return internal.ExternallyRenderContent(c.cfg, ctx, src, nodeBinary, args)
}

var nodeBinaryCandidates = []string{"node", "node.exe"}

func getNodeBinaryName() string {
	for _, candidate := range nodeBinaryCandidates {
		if hexec.InPath(candidate) {
			return candidate
		}
	}
	return ""
}

// Supports returns whether Markdoc is (or should be) installed on this computer.
func Supports() bool {
	hasNode := getNodeBinaryName() != ""
	// We also require the local script to be present; we can't easily check that here without FS.
	if htesting.SupportsAll() {
		if !hasNode {
			panic("node not installed")
		}
		return true
	}
	return hasNode
}
