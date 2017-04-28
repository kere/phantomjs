package phantomjs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	// ErrInjectionFailed is returned by InjectJS when injection fails.
	ErrInjectionFailed = errors.New("injection failed")
)

// Keyboard modifiers.
const (
	ShiftKey = 0x02000000
	CtrlKey  = 0x04000000
	AltKey   = 0x08000000
	MetaKey  = 0x10000000
	Keypad   = 0x20000000
)

// Default settings.
const (
	DefaultTimeOut = 10
	DefaultPort    = 20202
	DefaultBinPath = "phantomjs"
)

// Process represents a PhantomJS process.
type Process struct {
	TimeOut int
	Options []string
	path    string
	cmd     *exec.Cmd

	// Path to the 'phantomjs' binary.
	BinPath string

	// HTTP port used to communicate with phantomjs.
	Port int

	// Output from the process.
	Stdout io.Writer
	Stderr io.Writer
}

// NewProcess returns a new instance of Process.
func NewProcess() *Process {
	return &Process{
		TimeOut: DefaultTimeOut,
		BinPath: DefaultBinPath,
		Port:    DefaultPort,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
}

// Path returns a temporary path that the process is run from.
func (p *Process) Path() string {
	return p.path
}

// AddOption func
func (p *Process) AddOption(o string) {
	p.Options = append(p.Options, o)
}

// Open start the phantomjs process with the shim script.
func (p *Process) Open() error {
	if err := func() error {
		// Generate temporary path to run script from.
		// path, err := ioutil.TempDir("temp", "phantomjs-")
		// if err != nil {
		// 	return err
		// }
		path := filepath.Join(filepath.Dir("."), "temp")
		p.path = path
		if err := os.MkdirAll(path, os.ModeDir); err != nil {
			return err
		}

		// shim := shim1 + fmt.Sprint(p.Port) + shim2

		// Write shim script.
		// scriptPath := filepath.Join(path, fmt.Sprintf("shim.js", p.Port))
		scriptPath := filepath.Join(path, "shim.js")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			if err = ioutil.WriteFile(scriptPath, []byte(shim), 0600); err != nil {
				return err
			}
		}

		args := append([]string{scriptPath, fmt.Sprint(p.Port)}, p.Options...)

		// Start external process.
		cmd := exec.Command(p.BinPath, args...)
		// cmd.Env = []string{fmt.Sprintf("PORT=%d", p.Port)}
		cmd.Stdout = p.Stdout
		cmd.Stderr = p.Stderr
		if err := cmd.Start(); err != nil {
			return err
		}
		p.cmd = cmd

		// Wait until process is available.
		if err := p.wait(); err != nil {
			return err
		}
		return nil

	}(); err != nil {
		p.Close()
		return err
	}

	return nil
}

// Close stops the process.
func (p *Process) Close() (err error) {
	// Kill process.
	if p.cmd != nil {
		if e := p.cmd.Process.Kill(); e != nil && err == nil {
			err = e
		}
		p.cmd.Wait()
	}

	// Remove shim file.
	if p.path != "" {
		if e := os.RemoveAll(p.path); e != nil && err == nil {
			err = e
		}
	}

	return err
}

// URL returns the process' API URL.
func (p *Process) URL() string {
	return fmt.Sprintf("http://localhost:%d", p.Port)
}

// wait continually checks the process until it gets a response or times out.
func (p *Process) wait() error {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	timer := time.NewTimer(time.Duration(p.TimeOut) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			return errors.New("timeout")
		case <-ticker.C:
			if err := p.ping(); err == nil {
				return nil
			}
		}
	}
}

// ping checks the process to see if it is up.
func (p *Process) ping() error {
	// Send request.
	resp, err := http.Get(p.URL() + "/ping")
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Verify successful status code.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

// CreateWebPage returns a new instance of a "webpage".
func (p *Process) CreateWebPage() (*WebPage, error) {
	var resp struct {
		Ref refJSON `json:"ref"`
	}
	if err := p.doJSON("POST", "/webpage/Create", nil, &resp); err != nil {
		return nil, err
	}
	return &WebPage{ref: newRef(p, resp.Ref.ID)}, nil
}

// doJSON sends an HTTP request to url and encodes and decodes the req/resp as JSON.
func (p *Process) doJSON(method, path string, req, resp interface{}) error {
	// Encode request.
	var r io.Reader
	if req != nil {
		buf, err := json.Marshal(req)
		if err != nil {
			return err
		}
		r = bytes.NewReader(buf)
	}

	// Create request.
	httpRequest, err := http.NewRequest(method, p.URL()+path, r)
	if err != nil {
		return err
	}

	// Send request.
	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	// Read response body.
	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	// Check response code.
	if httpResponse.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found: %s", path)
	}

	// If an error was returned then return it.
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return errors.New("phantomjs.Process: " + string(body))
	} else if errResp.Error != "" {
		return errors.New(errResp.Error)
	}

	// Decode response if reference passed in.
	if resp != nil {
		if err := json.Unmarshal(body, resp); err != nil {
			return fmt.Errorf("unmarshal error: err=%s, body=%s", err, body)
		}
	}

	return nil
}

type errorResponse struct {
	Error string `json:"error"`
}

// DefaultProcess is a global, shared process.
// It must be opened before use.
var DefaultProcess = NewProcess()

// CreateWebPage returns a new instance of a "webpage" using the default process.
func CreateWebPage() (*WebPage, error) {
	return DefaultProcess.CreateWebPage()
}

// WebPage represents an object returned from "webpage.create()".
type WebPage struct {
	ref *Ref
}

// Open opens a URL.
func (p *WebPage) Open(url string) error {
	req := map[string]interface{}{
		"ref": p.ref.id,
		"url": url,
	}
	var resp struct {
		Status string `json:"status"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/Open", req, &resp); err != nil {
		return err
	}

	if resp.Status != "success" {
		return errors.New("failed")
	}
	return nil
}

// CanGoBack returns true if the page can be navigated back.
func (p *WebPage) CanGoBack() (bool, error) {
	var resp struct {
		Value bool `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/CanGoBack", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return false, err
	}
	return resp.Value, nil
}

// CanGoForward returns true if the page can be navigated forward.
func (p *WebPage) CanGoForward() (bool, error) {
	var resp struct {
		Value bool `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/CanGoForward", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return false, err
	}
	return resp.Value, nil
}

// ClipRect returns the clipping rectangle used when rendering.
// Returns nil if no clipping rectangle is set.
func (p *WebPage) ClipRect() (Rect, error) {
	var resp struct {
		Value rectJSON `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/ClipRect", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return Rect{}, err
	}
	return Rect{
		Top:    resp.Value.Top,
		Left:   resp.Value.Left,
		Width:  resp.Value.Width,
		Height: resp.Value.Height,
	}, nil
}

// SetClipRect sets the clipping rectangle used when rendering.
// Set to nil to render the entire webpage.
func (p *WebPage) SetClipRect(rect Rect) error {
	req := map[string]interface{}{
		"ref": p.ref.id,
		"rect": rectJSON{
			Top:    rect.Top,
			Left:   rect.Left,
			Width:  rect.Width,
			Height: rect.Height,
		},
	}
	return p.ref.process.doJSON("POST", "/webpage/SetClipRect", req, nil)
}

// Content returns content of the webpage enclosed in an HTML/XML element.
func (p *WebPage) Content() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/Content", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// SetContent sets the content of the webpage.
func (p *WebPage) SetContent(content string) error {
	return p.ref.process.doJSON("POST", "/webpage/SetContent", map[string]interface{}{"ref": p.ref.id, "content": content}, nil)
}

// Cookies returns a list of cookies visible to the current URL.
func (p *WebPage) Cookies() ([]*http.Cookie, error) {
	var resp struct {
		Value []cookieJSON `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/Cookies", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return nil, err
	}

	a := make([]*http.Cookie, len(resp.Value))
	for i := range resp.Value {
		a[i] = decodeCookieJSON(resp.Value[i])
	}
	return a, nil
}

// SetCookies sets a list of cookies visible to the current URL.
func (p *WebPage) SetCookies(cookies []*http.Cookie) error {
	a := make([]cookieJSON, len(cookies))
	for i := range cookies {
		a[i] = encodeCookieJSON(cookies[i])
	}
	req := map[string]interface{}{"ref": p.ref.id, "cookies": a}
	return p.ref.process.doJSON("POST", "/webpage/SetCookies", req, nil)
}

// CustomHeaders returns a list of additional headers sent with the web page.
func (p *WebPage) CustomHeaders() (http.Header, error) {
	var resp struct {
		Value map[string]string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/CustomHeaders", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return nil, err
	}

	// Convert to a header object.
	hdr := make(http.Header)
	for key, value := range resp.Value {
		hdr.Set(key, value)
	}
	return hdr, nil
}

// SetCustomHeaders sets a list of additional headers sent with the web page.
//
// This function does not support multiple headers with the same name. Only
// the first value for a header key will be used.
func (p *WebPage) SetCustomHeaders(header http.Header) error {
	m := make(map[string]string)
	for key := range header {
		m[key] = header.Get(key)
	}
	req := map[string]interface{}{"ref": p.ref.id, "headers": m}
	return p.ref.process.doJSON("POST", "/webpage/SetCustomHeaders", req, nil)
}

// FocusedFrameName returns the name of the currently focused frame.
func (p *WebPage) FocusedFrameName() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/FocusedFrameName", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// FrameContent returns the content of the current frame.
func (p *WebPage) FrameContent() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/FrameContent", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// SetFrameContent sets the content of the current frame.
func (p *WebPage) SetFrameContent(content string) error {
	return p.ref.process.doJSON("POST", "/webpage/SetFrameContent", map[string]interface{}{"ref": p.ref.id, "content": content}, nil)
}

// FrameName returns the name of the current frame.
func (p *WebPage) FrameName() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/FrameName", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// FramePlainText returns the plain text representation of the current frame content.
func (p *WebPage) FramePlainText() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/FramePlainText", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// FrameTitle returns the title of the current frame.
func (p *WebPage) FrameTitle() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/FrameTitle", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// FrameURL returns the URL of the current frame.
func (p *WebPage) FrameURL() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/FrameURL", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// FrameCount returns the total number of frames.
func (p *WebPage) FrameCount() (int, error) {
	var resp struct {
		Value int `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/FrameCount", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return 0, err
	}
	return resp.Value, nil
}

// FrameNames returns an list of frame names.
func (p *WebPage) FrameNames() ([]string, error) {
	var resp struct {
		Value []string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/FrameNames", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// LibraryPath returns the path used by InjectJS() to resolve scripts.
// Initially it is set to Process.Path().
func (p *WebPage) LibraryPath() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/LibraryPath", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// SetLibraryPath sets the library path used by InjectJS().
func (p *WebPage) SetLibraryPath(path string) error {
	return p.ref.process.doJSON("POST", "/webpage/SetLibraryPath", map[string]interface{}{"ref": p.ref.id, "path": path}, nil)
}

// NavigationLocked returns true if the navigation away from the page is disabled.
func (p *WebPage) NavigationLocked() (bool, error) {
	var resp struct {
		Value bool `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/NavigationLocked", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return false, err
	}
	return resp.Value, nil
}

// SetNavigationLocked sets whether navigation away from the page should be disabled.
func (p *WebPage) SetNavigationLocked(value bool) error {
	return p.ref.process.doJSON("POST", "/webpage/SetNavigationLocked", map[string]interface{}{"ref": p.ref.id, "value": value}, nil)
}

// OfflineStoragePath returns the path used by offline storage.
func (p *WebPage) OfflineStoragePath() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/OfflineStoragePath", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// OfflineStorageQuota returns the number of bytes that can be used for offline storage.
func (p *WebPage) OfflineStorageQuota() (int, error) {
	var resp struct {
		Value int `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/OfflineStorageQuota", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return 0, err
	}
	return resp.Value, nil
}

// OwnsPages returns true if this page owns pages opened in other windows.
func (p *WebPage) OwnsPages() (bool, error) {
	var resp struct {
		Value bool `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/OwnsPages", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return false, err
	}
	return resp.Value, nil
}

// SetOwnsPages sets whether this page owns pages opened in other windows.
func (p *WebPage) SetOwnsPages(v bool) error {
	return p.ref.process.doJSON("POST", "/webpage/SetOwnsPages", map[string]interface{}{"ref": p.ref.id, "value": v}, nil)
}

// PageWindowNames returns an list of owned window names.
func (p *WebPage) PageWindowNames() ([]string, error) {
	var resp struct {
		Value []string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/PageWindowNames", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return nil, err
	}
	return resp.Value, nil
}

// Pages returns a list of owned pages.
func (p *WebPage) Pages() ([]*WebPage, error) {
	var resp struct {
		Refs []refJSON `json:"refs"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/Pages", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return nil, err
	}

	// Convert reference IDs to web pages.
	a := make([]*WebPage, len(resp.Refs))
	for i, ref := range resp.Refs {
		a[i] = &WebPage{ref: newRef(p.ref.process, ref.ID)}
	}
	return a, nil
}

// PaperSize returns the size of the web page when rendered as a PDF.
func (p *WebPage) PaperSize() (PaperSize, error) {
	var resp struct {
		Value paperSizeJSON `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/PaperSize", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return PaperSize{}, err
	}
	return decodePaperSizeJSON(resp.Value), nil
}

// SetPaperSize sets the size of the web page when rendered as a PDF.
func (p *WebPage) SetPaperSize(size PaperSize) error {
	req := map[string]interface{}{"ref": p.ref.id, "size": encodePaperSizeJSON(size)}
	return p.ref.process.doJSON("POST", "/webpage/SetPaperSize", req, nil)
}

// PlainText returns the plain text representation of the page.
func (p *WebPage) PlainText() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/PlainText", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// ScrollPosition returns the current scroll position of the page.
func (p *WebPage) ScrollPosition() (Position, error) {
	var resp struct {
		Top  int `json:"top"`
		Left int `json:"left"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/ScrollPosition", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return Position{}, err
	}
	return Position{Top: resp.Top, Left: resp.Left}, nil
}

// SetScrollPosition sets the current scroll position of the page.
func (p *WebPage) SetScrollPosition(pos Position) error {
	return p.ref.process.doJSON("POST", "/webpage/SetScrollPosition", map[string]interface{}{"ref": p.ref.id, "top": pos.Top, "left": pos.Left}, nil)
}

// Settings returns the settings used on the web page.
func (p *WebPage) Settings() (WebPageSettings, error) {
	var resp struct {
		Settings webPageSettingsJSON `json:"settings"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/Settings", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return WebPageSettings{}, err
	}
	return WebPageSettings{
		JavascriptEnabled:             resp.Settings.JavascriptEnabled,
		LoadImages:                    resp.Settings.LoadImages,
		LocalToRemoteURLAccessEnabled: resp.Settings.LocalToRemoteURLAccessEnabled,
		UserAgent:                     resp.Settings.UserAgent,
		Username:                      resp.Settings.Username,
		Password:                      resp.Settings.Password,
		XSSAuditingEnabled:            resp.Settings.XSSAuditingEnabled,
		WebSecurityEnabled:            resp.Settings.WebSecurityEnabled,
		ResourceTimeout:               time.Duration(resp.Settings.ResourceTimeout) * time.Millisecond,
	}, nil
}

// SetSettings sets various settings on the web page.
//
// The settings apply only during the initial call to the page.open function.
// Subsequent modification of the settings object will not have any impact.
func (p *WebPage) SetSettings(settings WebPageSettings) error {
	req := map[string]interface{}{
		"ref": p.ref.id,
		"settings": webPageSettingsJSON{
			JavascriptEnabled:             settings.JavascriptEnabled,
			LoadImages:                    settings.LoadImages,
			LocalToRemoteURLAccessEnabled: settings.LocalToRemoteURLAccessEnabled,
			UserAgent:                     settings.UserAgent,
			Username:                      settings.Username,
			Password:                      settings.Password,
			XSSAuditingEnabled:            settings.XSSAuditingEnabled,
			WebSecurityEnabled:            settings.WebSecurityEnabled,
			ResourceTimeout:               int(settings.ResourceTimeout / time.Millisecond),
		},
	}
	return p.ref.process.doJSON("POST", "/webpage/SetSettings", req, nil)
}

// Title returns the title of the web page.
func (p *WebPage) Title() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/Title", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// URL returns the current URL of the web page.
func (p *WebPage) URL() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/URL", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// ViewportSize returns the size of the viewport on the browser.
func (p *WebPage) ViewportSize() (width, height int, err error) {
	var resp struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/ViewportSize", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return 0, 0, err
	}
	return resp.Width, resp.Height, nil
}

// SetViewportSize sets the size of the viewport.
func (p *WebPage) SetViewportSize(width, height int) error {
	return p.ref.process.doJSON("POST", "/webpage/SetViewportSize", map[string]interface{}{"ref": p.ref.id, "width": width, "height": height}, nil)
}

// WindowName returns the window name of the web page.
func (p *WebPage) WindowName() (string, error) {
	var resp struct {
		Value string `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/WindowName", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

// ZoomFactor returns zoom factor when rendering the page.
func (p *WebPage) ZoomFactor() (float64, error) {
	var resp struct {
		Value float64 `json:"value"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/ZoomFactor", map[string]interface{}{"ref": p.ref.id}, &resp); err != nil {
		return 0, err
	}
	return resp.Value, nil
}

// SetZoomFactor sets the zoom factor when rendering the page.
func (p *WebPage) SetZoomFactor(factor float64) error {
	return p.ref.process.doJSON("POST", "/webpage/SetZoomFactor", map[string]interface{}{"ref": p.ref.id, "value": factor}, nil)
}

// AddCookie adds a cookie to the page.
// Returns true if the cookie was successfully added.
func (p *WebPage) AddCookie(cookie *http.Cookie) (bool, error) {
	var resp struct {
		ReturnValue bool `json:"returnValue"`
	}
	req := map[string]interface{}{"ref": p.ref.id, "cookie": encodeCookieJSON(cookie)}
	if err := p.ref.process.doJSON("POST", "/webpage/AddCookie", req, &resp); err != nil {
		return false, err
	}
	return resp.ReturnValue, nil
}

// ClearCookies deletes all cookies visible to the current URL.
func (p *WebPage) ClearCookies() error {
	return p.ref.process.doJSON("POST", "/webpage/ClearCookies", map[string]interface{}{"ref": p.ref.id}, nil)
}

// Close releases the web page and its resources.
func (p *WebPage) Close() error {
	return p.ref.process.doJSON("POST", "/webpage/Close", map[string]interface{}{"ref": p.ref.id}, nil)
}

// DeleteCookie removes a cookie with a matching name.
// Returns true if the cookie was successfully deleted.
func (p *WebPage) DeleteCookie(name string) (bool, error) {
	var resp struct {
		ReturnValue bool `json:"returnValue"`
	}
	req := map[string]interface{}{"ref": p.ref.id, "name": name}
	if err := p.ref.process.doJSON("POST", "/webpage/DeleteCookie", req, &resp); err != nil {
		return false, err
	}
	return resp.ReturnValue, nil
}

// EvaluateAsync executes a JavaScript function and returns immediately.
// Execution is delayed by delay. No value is returned.
func (p *WebPage) EvaluateAsync(script string, delay time.Duration) error {
	return p.ref.process.doJSON("POST", "/webpage/EvaluateAsync", map[string]interface{}{"ref": p.ref.id, "script": script, "delay": int(delay / time.Millisecond)}, nil)
}

// EvaluateJavaScript executes a JavaScript function.
// Returns the value returned by the function.
func (p *WebPage) EvaluateJavaScript(script string) (interface{}, error) {
	var resp struct {
		ReturnValue interface{} `json:"returnValue"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/EvaluateJavaScript", map[string]interface{}{"ref": p.ref.id, "script": script}, &resp); err != nil {
		return nil, err
	}
	return resp.ReturnValue, nil
}

// Evaluate executes a JavaScript function in the context of the web page.
// Returns the value returned by the function.
func (p *WebPage) Evaluate(script string) (interface{}, error) {
	var resp struct {
		ReturnValue interface{} `json:"returnValue"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/Evaluate", map[string]interface{}{"ref": p.ref.id, "script": script}, &resp); err != nil {
		return nil, err
	}
	return resp.ReturnValue, nil
}

// Page returns an owned page by window name.
// Returns nil if the page cannot be found.
func (p *WebPage) Page(name string) (*WebPage, error) {
	var resp struct {
		Ref refJSON `json:"ref"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/Page", map[string]interface{}{"ref": p.ref.id, "name": name}, &resp); err != nil {
		return nil, err
	}
	if resp.Ref.ID == "" {
		return nil, nil
	}
	return &WebPage{ref: newRef(p.ref.process, resp.Ref.ID)}, nil
}

// GoBack navigates back to the previous page.
func (p *WebPage) GoBack() error {
	return p.ref.process.doJSON("POST", "/webpage/GoBack", map[string]interface{}{"ref": p.ref.id}, nil)
}

// GoForward navigates to the next page.
func (p *WebPage) GoForward() error {
	return p.ref.process.doJSON("POST", "/webpage/GoForward", map[string]interface{}{"ref": p.ref.id}, nil)
}

// Go navigates to the page in history by relative offset.
// A positive index moves forward, a negative index moves backwards.
func (p *WebPage) Go(index int) error {
	return p.ref.process.doJSON("POST", "/webpage/Go", map[string]interface{}{"ref": p.ref.id, "index": index}, nil)
}

// IncludeJS includes an external script from url.
// Returns after the script has been loaded.
func (p *WebPage) IncludeJS(url string) error {
	return p.ref.process.doJSON("POST", "/webpage/IncludeJS", map[string]interface{}{"ref": p.ref.id, "url": url}, nil)
}

// InjectJS injects an external script from the local filesystem.
//
// The script will be loaded from the Process.Path() directory. If it cannot be
// found then it is loaded from the library path.
func (p *WebPage) InjectJS(filename string) error {
	var resp struct {
		ReturnValue bool `json:"returnValue"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/InjectJS", map[string]interface{}{"ref": p.ref.id, "filename": filename}, &resp); err != nil {
		return err
	}
	if !resp.ReturnValue {
		return ErrInjectionFailed
	}
	return nil
}

// Reload reloads the current web page.
func (p *WebPage) Reload() error {
	return p.ref.process.doJSON("POST", "/webpage/Reload", map[string]interface{}{"ref": p.ref.id}, nil)
}

// RenderBase64 renders the web page to a base64 encoded string.
func (p *WebPage) RenderBase64(format string) (string, error) {
	var resp struct {
		ReturnValue string `json:"returnValue"`
	}
	if err := p.ref.process.doJSON("POST", "/webpage/RenderBase64", map[string]interface{}{"ref": p.ref.id, "format": format}, &resp); err != nil {
		return "", err
	}
	return resp.ReturnValue, nil
}

// Render renders the web page to a file with the given format and quality settings.
// This supports the "PDF", "PNG", "JPEG", "BMP", "PPM", and "GIF" formats.
func (p *WebPage) Render(filename, format string, quality int) error {
	req := map[string]interface{}{"ref": p.ref.id, "filename": filename, "format": format, "quality": quality}
	return p.ref.process.doJSON("POST", "/webpage/Render", req, nil)
}

// SendMouseEvent sends a mouse event as if it came from the user.
// It is not a synthetic event.
//
// The eventType can be "mouseup", "mousedown", "mousemove", "doubleclick",
// or "click". The mouseX and mouseY specify the position of the mouse on the
// screen. The button argument specifies the mouse button clicked (e.g. "left").
func (p *WebPage) SendMouseEvent(eventType string, mouseX, mouseY int, button string) error {
	return p.ref.process.doJSON("POST", "/webpage/SendMouseEvent", map[string]interface{}{"ref": p.ref.id, "eventType": eventType, "mouseX": mouseX, "mouseY": mouseY, "button": button}, nil)
}

// SendKeyboardEvent sends a keyboard event as if it came from the user.
// It is not a synthetic event.
//
// The eventType can be "keyup", "keypress", or "keydown".
//
// The key argument is a string or a key listed here:
// https://github.com/ariya/phantomjs/commit/cab2635e66d74b7e665c44400b8b20a8f225153a
//
// Keyboard modifiers can be joined together using the bitwise OR operator.
func (p *WebPage) SendKeyboardEvent(eventType string, key string, modifier int) error {
	return p.ref.process.doJSON("POST", "/webpage/SendKeyboardEvent", map[string]interface{}{"ref": p.ref.id, "eventType": eventType, "key": key, "modifier": modifier}, nil)
}

// SetContentAndURL sets the content and URL of the page.
func (p *WebPage) SetContentAndURL(content, url string) error {
	return p.ref.process.doJSON("POST", "/webpage/SetContentAndURL", map[string]interface{}{"ref": p.ref.id, "content": content, "url": url}, nil)
}

// Stop stops the web page.
func (p *WebPage) Stop() error {
	return p.ref.process.doJSON("POST", "/webpage/Stop", map[string]interface{}{"ref": p.ref.id}, nil)
}

// SwitchToFocusedFrame changes the current frame to the frame that is in focus.
func (p *WebPage) SwitchToFocusedFrame() error {
	return p.ref.process.doJSON("POST", "/webpage/SwitchToFocusedFrame", map[string]interface{}{"ref": p.ref.id}, nil)
}

// SwitchToFrameName changes the current frame to a frame with a given name.
func (p *WebPage) SwitchToFrameName(name string) error {
	return p.ref.process.doJSON("POST", "/webpage/SwitchToFrameName", map[string]interface{}{"ref": p.ref.id, "name": name}, nil)
}

// SwitchToFramePosition changes the current frame to the frame at the given position.
func (p *WebPage) SwitchToFramePosition(pos int) error {
	return p.ref.process.doJSON("POST", "/webpage/SwitchToFramePosition", map[string]interface{}{"ref": p.ref.id, "position": pos}, nil)
}

// SwitchToMainFrame switches the current frame to the main frame.
func (p *WebPage) SwitchToMainFrame() error {
	return p.ref.process.doJSON("POST", "/webpage/SwitchToMainFrame", map[string]interface{}{"ref": p.ref.id}, nil)
}

// SwitchToParentFrame switches the current frame to the parent of the current frame.
func (p *WebPage) SwitchToParentFrame() error {
	return p.ref.process.doJSON("POST", "/webpage/SwitchToParentFrame", map[string]interface{}{"ref": p.ref.id}, nil)
}

// UploadFile uploads a file to a form element specified by selector.
func (p *WebPage) UploadFile(selector, filename string) error {
	return p.ref.process.doJSON("POST", "/webpage/UploadFile", map[string]interface{}{"ref": p.ref.id, "selector": selector, "filename": filename}, nil)
}

// OpenWebPageSettings represents the settings object passed to WebPage.Open().
type OpenWebPageSettings struct {
	Method string `json:"method"`
}

// Ref represents a reference to an object in phantomjs.
type Ref struct {
	process *Process
	id      string
}

// newRef returns a new instance of a referenced object within the process.
func newRef(p *Process, id string) *Ref {
	return &Ref{process: p, id: id}
}

// ID returns the reference identifier.
func (r *Ref) ID() string {
	return r.id
}

// refJSON is a struct for encoding refs as JSON.
type refJSON struct {
	ID string `json:"id"`
}

// Rect represents a rectangle used by WebPage.ClipRect().
type Rect struct {
	Top    int
	Left   int
	Width  int
	Height int
}

// rectJSON is a struct for encoding rects as JSON.
type rectJSON struct {
	Top    int `json:"top"`
	Left   int `json:"left"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// cookieJSON is a struct for encoding http.Cookie objects as JSON.
type cookieJSON struct {
	Domain   string `json:"domain"`
	Expires  string `json:"expires"`
	Expiry   int    `json:"expiry"`
	HTTPOnly bool   `json:"httponly"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Secure   bool   `json:"secure"`
	Value    string `json:"value"`
}

func encodeCookieJSON(v *http.Cookie) cookieJSON {
	out := cookieJSON{
		Domain:   v.Domain,
		HTTPOnly: v.HttpOnly,
		Name:     v.Name,
		Path:     v.Path,
		Secure:   v.Secure,
		Value:    v.Value,
	}

	if !v.Expires.IsZero() {
		out.Expires = v.Expires.UTC().Format(http.TimeFormat)
	}
	return out
}

func decodeCookieJSON(v cookieJSON) *http.Cookie {
	out := &http.Cookie{
		Domain:     v.Domain,
		RawExpires: v.Expires,
		HttpOnly:   v.HTTPOnly,
		Name:       v.Name,
		Path:       v.Path,
		Secure:     v.Secure,
		Value:      v.Value,
	}

	if v.Expires != "" {
		expires, _ := time.Parse(http.TimeFormat, v.Expires)
		out.Expires = expires
		out.RawExpires = v.Expires
	}

	return out
}

// PaperSize represents the size of a webpage when rendered as a PDF.
//
// Units can be specified in "mm", "cm", "in", or "px".
// If no unit is specified then "px" is used.
type PaperSize struct {
	// Dimensions of the paper.
	// This can also be specified via Format.
	Width  string
	Height string

	// Supported formats: "A3", "A4", "A5", "Legal", "Letter", "Tabloid".
	Format string

	// Margins around the paper.
	Margin *PaperSizeMargin

	// Supported orientations: "portrait", "landscape".
	Orientation string
}

// PaperSizeMargin represents the margins around the paper.
type PaperSizeMargin struct {
	Top    string
	Bottom string
	Left   string
	Right  string
}

type paperSizeJSON struct {
	Width       string               `json:"width,omitempty"`
	Height      string               `json:"height,omitempty"`
	Format      string               `json:"format,omitempty"`
	Margin      *paperSizeMarginJSON `json:"margin,omitempty"`
	Orientation string               `json:"orientation,omitempty"`
}

type paperSizeMarginJSON struct {
	Top    string `json:"top,omitempty"`
	Bottom string `json:"bottom,omitempty"`
	Left   string `json:"left,omitempty"`
	Right  string `json:"right,omitempty"`
}

func encodePaperSizeJSON(v PaperSize) paperSizeJSON {
	out := paperSizeJSON{
		Width:       v.Width,
		Height:      v.Height,
		Format:      v.Format,
		Orientation: v.Orientation,
	}
	if v.Margin != nil {
		out.Margin = &paperSizeMarginJSON{
			Top:    v.Margin.Top,
			Bottom: v.Margin.Bottom,
			Left:   v.Margin.Left,
			Right:  v.Margin.Right,
		}
	}
	return out
}

func decodePaperSizeJSON(v paperSizeJSON) PaperSize {
	out := PaperSize{
		Width:       v.Width,
		Height:      v.Height,
		Format:      v.Format,
		Orientation: v.Orientation,
	}
	if v.Margin != nil {
		out.Margin = &PaperSizeMargin{
			Top:    v.Margin.Top,
			Bottom: v.Margin.Bottom,
			Left:   v.Margin.Left,
			Right:  v.Margin.Right,
		}
	}
	return out
}

// Position represents a coordinate on the page, in pixels.
type Position struct {
	Top  int
	Left int
}

// WebPageSettings represents various settings on a web page.
type WebPageSettings struct {
	JavascriptEnabled             bool
	LoadImages                    bool
	LocalToRemoteURLAccessEnabled bool
	UserAgent                     string
	Username                      string
	Password                      string
	XSSAuditingEnabled            bool
	WebSecurityEnabled            bool
	ResourceTimeout               time.Duration
}

type webPageSettingsJSON struct {
	JavascriptEnabled             bool   `json:"javascriptEnabled"`
	LoadImages                    bool   `json:"loadImages"`
	LocalToRemoteURLAccessEnabled bool   `json:"localToRemoteUrlAccessEnabled"`
	UserAgent                     string `json:"userAgent"`
	Username                      string `json:"username"`
	Password                      string `json:"password"`
	XSSAuditingEnabled            bool   `json:"XSSAuditingEnabled"`
	WebSecurityEnabled            bool   `json:"webSecurityEnabled"`
	ResourceTimeout               int    `json:"resourceTimeout"`
}
