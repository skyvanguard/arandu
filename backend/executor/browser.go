package executor

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/arandu-ai/arandu/assets"
	"github.com/arandu-ai/arandu/config"
	"github.com/arandu-ai/arandu/database"
	"github.com/arandu-ai/arandu/logging"
	"github.com/arandu-ai/arandu/templates"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const port = "9222"

func InitBrowser(db *database.Queries) error {
	browserContainerName := BrowserName()
	portBinding := nat.Port(fmt.Sprintf("%s/tcp", port))

	_, err := SpawnContainer(context.Background(), browserContainerName, &container.Config{
		Image: "ghcr.io/go-rod/rod",
		ExposedPorts: nat.PortSet{
			portBinding: struct{}{},
		},
		Cmd: []string{"chrome", "--headless", "--no-sandbox", fmt.Sprintf("--remote-debugging-port=%s", port), "--remote-debugging-address=0.0.0.0"},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			portBinding: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
	}, db)

	if err != nil {
		return fmt.Errorf("failed to spawn container: %w", err)
	}

	return nil
}

func Content(url string) (result string, screenshotName string, err error) {
	logging.Debug("Trying to get content from URL", "url", url)

	page, err := loadPage()

	if err != nil {
		return "", "", fmt.Errorf("error loading page: %w", err)
	}

	err = loadUrl(page, url)

	if err != nil {
		return "", "", fmt.Errorf("error loading url: %w", err)
	}

	script, err := templates.Render(assets.ScriptTemplates, "scripts/content.js", nil)

	if err != nil {
		return "", "", fmt.Errorf("error reading script: %w", err)
	}

	pageText, err := page.Eval(string(script))

	if err != nil {
		return "", "", fmt.Errorf("error evaluating script: %w", err)
	}

	screenshot, err := page.Screenshot(false, nil)

	if err != nil {
		return "", "", fmt.Errorf("error taking screenshot: %w", err)
	}

	screenshotName, err = writeScreenshotToFile(screenshot)

	if err != nil {
		return "", "", fmt.Errorf("error writing screenshot to file: %w", err)
	}

	return pageText.Value.Str(), screenshotName, nil
}

func URLs(url string) (result string, screenshotName string, err error) {
	logging.Debug("Trying to get URLs from page", "url", url)

	page, err := loadPage()

	if err != nil {
		return "", "", fmt.Errorf("error loading page: %w", err)
	}

	err = loadUrl(page, url)

	if err != nil {
		return "", "", fmt.Errorf("error loading url: %w", err)
	}

	script, err := templates.Render(assets.ScriptTemplates, "scripts/urls.js", nil)

	if err != nil {
		return "", "", fmt.Errorf("error reading script: %w", err)
	}

	urls, err := page.Eval(string(script))

	if err != nil {
		return "", "", fmt.Errorf("error evaluating script: %w", err)
	}

	screenshot, err := page.Screenshot(true, nil)

	if err != nil {
		return "", "", fmt.Errorf("error taking screenshot: %w", err)
	}

	screenshotName, err = writeScreenshotToFile(screenshot)

	if err != nil {
		return "", "", fmt.Errorf("error writing screenshot to file: %w", err)
	}

	return urls.Value.Str(), screenshotName, nil
}

func writeScreenshotToFile(screenshot []byte) (filename string, err error) {
	// Write screenshot to file
	filename = fmt.Sprintf("%s.png", time.Now().Format("2006-01-02-15-04-05"))
	path := "./tmp/browser/"
	filepath := fmt.Sprintf("./tmp/browser/%s", filename)

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("error creating directory: %w", err)
	}

	file, err := os.Create(filepath)

	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}

	defer file.Close()

	_, err = file.Write(screenshot)

	if err != nil {
		return "", fmt.Errorf("error writing to file: %w", err)
	}

	return filename, nil
}

func BrowserName() string {
	return "arandu-browser"
}

func loadPage() (*rod.Page, error) {
	// Bug fix #65: Use configurable Chrome debug URL or try multiple fallbacks
	var u string
	var err error

	if config.Config.ChromeDebugURL != "" {
		// Use configured URL
		u = config.Config.ChromeDebugURL
		logging.Debug("Using configured Chrome debug URL", "url", u)
	} else {
		// Try to resolve using launcher (works when Chrome is on localhost)
		u, err = launcher.ResolveURL("")
		if err != nil {
			// Fallback: try connecting to the browser container by name
			// This works in Docker networking where containers can reach each other by name
			u = fmt.Sprintf("ws://%s:%s", BrowserName(), port)
			logging.Debug("Fallback to container name", "url", u)
		}
	}

	browser := rod.New().ControlURL(u)
	err = browser.Connect()

	if err != nil {
		// If connection fails, try alternative URLs
		alternativeURLs := []string{
			fmt.Sprintf("ws://localhost:%s", port),
			fmt.Sprintf("ws://127.0.0.1:%s", port),
			fmt.Sprintf("ws://host.docker.internal:%s", port),
			fmt.Sprintf("ws://%s:%s", BrowserName(), port),
		}

		for _, altURL := range alternativeURLs {
			logging.Debug("Trying alternative Chrome URL", "url", altURL)
			browser = rod.New().ControlURL(altURL)
			err = browser.Connect()
			if err == nil {
				break
			}
		}

		if err != nil {
			return nil, fmt.Errorf("error connecting to browser (tried multiple URLs): %w", err)
		}
	}

	version, err := browser.Version()

	if err != nil {
		return nil, fmt.Errorf("error getting browser version: %w", err)
	}
	logging.Info("Connected to browser", "product", version.Product)

	page, err := browser.Page(proto.TargetCreateTarget{})

	if err != nil {
		return nil, fmt.Errorf("error opening page: %w", err)
	}

	return page, nil
}

func loadUrl(page *rod.Page, url string) error {
	pageRouter := page.HijackRequests()

	// Do not load any images or css files
	pageRouter.MustAdd("*", func(ctx *rod.Hijack) {
		// There're a lot of types you can use in this enum, like NetworkResourceTypeScript for javascript files
		// In this case we're using NetworkResourceTypeImage to block images
		if ctx.Request.Type() == proto.NetworkResourceTypeImage ||
			ctx.Request.Type() == proto.NetworkResourceTypeStylesheet ||
			ctx.Request.Type() == proto.NetworkResourceTypeFont ||
			ctx.Request.Type() == proto.NetworkResourceTypeMedia ||
			ctx.Request.Type() == proto.NetworkResourceTypeManifest ||
			ctx.Request.Type() == proto.NetworkResourceTypeOther {
			ctx.Response.Fail(proto.NetworkErrorReasonBlockedByClient)
			return
		}
		ctx.ContinueRequest(&proto.FetchContinueRequest{})
	})

	// since we are only hijacking a specific page, even using the "*" won't affect much of the performance
	go pageRouter.Run()

	err := page.Navigate(url)

	if err != nil {
		return fmt.Errorf("error navigating to page: %w", err)
	}

	err = page.WaitDOMStable(time.Second*1, 5)

	if err != nil {
		return fmt.Errorf("error waiting for page to stabilize: %w", err)
	}

	return nil
}
