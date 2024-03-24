package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	rerrros "github.com/container-registry/helm-charts-oci-proxy/registry/errors"
	"github.com/container-registry/helm-charts-oci-proxy/registry/models"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"net/http"
	"net/url"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
	"path"
	"strings"
)

const (
	DBName         = "registry"
	DefaultTagName = "latest"
)

type Registry struct {
	log   logrus.StdLogger
	debug bool
	//
	ociRegistry string
	//
	ociUsername   string
	ociPassword   string
	ociHostname   string
	ociRepository string
	initialised   bool
}

// Option describes the available options
// for creating the registry.
type Option func(r *Registry)

func New(opts ...Option) *Registry {
	r := &Registry{}
	for _, o := range opts {
		o(r)
	}
	if r.log == nil {
		r.log = log.Default()
	}
	return r
}

func WithOciURL(url string) Option {
	return func(r *Registry) {
		r.ociRegistry = url
	}
}

func WithDebug(enable bool) Option {
	return func(r *Registry) {
		r.debug = enable
	}
}

func WithLogger(l logrus.StdLogger) Option {
	return func(r *Registry) {
		r.log = l
	}
}

func (r *Registry) handle(resp http.ResponseWriter, req *http.Request) error {

	if !r.initialised {
		return fmt.Errorf("not initialised")
	}
	path := req.URL.Path

	if path == "/" || path == "" {
		return r.homeHandler(resp, req)
	}

	if req.Method == http.MethodPut && strings.HasPrefix(path, "/-/user/") {
		return r.loginHandler(resp, req)
	}

	if req.Method == http.MethodPut {
		return r.pushPackage(resp, req)
	}

	if req.Method == http.MethodGet {
		return r.pullPackage(resp, req)
	}

	return rerrros.RegErrNotFound
}

func (r *Registry) pushPackage(resp http.ResponseWriter, req *http.Request) error {

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	var result models.Package
	if err = json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		return err
	}

	if len(result.Attachments) == 0 {
		return fmt.Errorf("no attachments")
	}

	if len(result.Versions) == 0 {
		return fmt.Errorf("no versions")
	}

	if err = r.pushToRegistry(req.Context(), result); err != nil {
		return err
	}
	resp.WriteHeader(http.StatusOK)
	return nil
}

func (r *Registry) homeHandler(resp http.ResponseWriter, req *http.Request) error {
	//
	resp.WriteHeader(200)
	if err := prettyEncode(&models.Registry{
		DbName: DBName,
	}, resp); err != nil {
		return rerrros.RegErrInternal(err)
	}
	return nil
}

// loginHandler @TODO implement the real authorisation flow
func (r *Registry) loginHandler(resp http.ResponseWriter, req *http.Request) error {
	resp.WriteHeader(200)
	if err := prettyEncode(&models.LoginResponse{
		OK:    true,
		Token: "dummy-token",
		ID:    fmt.Sprintf("org.couchdb.user:undefined"),
	}, resp); err != nil {
		return rerrros.RegErrInternal(err)
	}
	return nil
}

func (r *Registry) Handle(resp http.ResponseWriter, req *http.Request) {
	if r.debug {
		r.log.Printf("%s - %s", req.Method, req.URL)
	}
	if err := r.handle(resp, req); err != nil {
		var regErr *rerrros.RegError
		if errors.As(err, &regErr) {
			r.log.Printf("%s %s %d %s %s", req.Method, req.URL, regErr.Status, regErr.Code, regErr.Message)
			_ = regErr.Write(resp)
			return
		}
		r.log.Printf("%s %s %s", req.Method, req.URL, err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (r *Registry) pullPackage(resp http.ResponseWriter, req *http.Request) error {

	if req.URL == nil {
		return fmt.Errorf("internal error")
	}
	ctx := req.Context()

	src, err := remote.NewRepository(r.ociHostname + strings.TrimSuffix(r.ociRepository, "/") + "/" + path.Base(req.URL.Path))
	if err != nil {
		return err
	}

	src.PlainHTTP = true
	dst := memory.New()

	tagName := DefaultTagName

	desc, err := oras.Copy(ctx, src, tagName, dst, tagName, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	readCloser, err := dst.Fetch(ctx, desc)
	if err != nil {
		return err
	}
	var manifest v1.Manifest

	if err := json.NewDecoder(readCloser).Decode(&manifest); err != nil {
		return err
	}

	dataReadCloser, err := dst.Fetch(ctx, manifest.Config)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(dataReadCloser)
	if err != nil {
		return err
	}
	resp.WriteHeader(http.StatusOK)
	resp.Write(data)
	return nil
}

func pushBlob(ctx context.Context, mediaType string, blob []byte, target oras.Target) (desc v1.Descriptor, err error) {
	desc = v1.Descriptor{ // Generate descriptor based on the media type and blob content
		MediaType: mediaType,
		Digest:    digest.FromBytes(blob), // Calculate digest
		Size:      int64(len(blob)),       // Include blob size
	}
	return desc, target.Push(ctx, desc, bytes.NewReader(blob)) // Push the blob to the registry target
}

func generateManifestContent(config v1.Descriptor, layers ...v1.Descriptor) ([]byte, error) {
	content := v1.Manifest{
		Config:    config, // Set config blob
		Layers:    layers, // Set layer blobs
		Versioned: specs.Versioned{SchemaVersion: 2},
	}
	return json.Marshal(content) // Get json content
}

func (r *Registry) pushToRegistry(ctx context.Context, pkg models.Package) error {

	mem := memory.New()
	data, err := json.Marshal(pkg)
	if err != nil {
		return err
	}

	layerDesc, err := pushBlob(ctx, v1.MediaTypeImageLayer, data, mem) // push layer blob
	if err != nil {
		return err
	}
	manifestBlob, err := generateManifestContent(layerDesc) // generate a image manifest
	if err != nil {
		return err
	}
	manifestDesc, err := pushBlob(ctx, v1.MediaTypeImageManifest, manifestBlob, mem) // push manifest blob
	if err != nil {
		return err
	}

	tag := DefaultTagName
	err = mem.Tag(ctx, manifestDesc, tag)
	if err != nil {
		return err
	}

	// 3. Connect to a remote repository
	repo, err := remote.NewRepository(r.ociHostname + strings.TrimSuffix(r.ociRepository, "/") + "/" + path.Base(pkg.Name))
	if err != nil {
		return err
	}
	repo.PlainHTTP = true

	// Note: The below code can be omitted if authentication is not required
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(r.ociHostname, auth.Credential{
			Username: r.ociUsername,
			Password: r.ociPassword,
		}),
	}

	// 4. Copy from the file store to the remote repository
	_, err = oras.Copy(ctx, mem, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	return nil
}

// Init tests if configuration is correct
// apply
func (r *Registry) Init() error {

	u, err := url.Parse(r.ociRegistry)
	if err != nil {
		return err
	}
	r.ociHostname = u.Hostname()
	if u.User != nil {
		r.ociUsername = u.User.Username()
		r.ociPassword, _ = u.User.Password()
	}
	r.ociRepository = u.Path
	if r.ociRepository == "" {
		r.ociRepository = "/"
	}
	r.initialised = true

	return nil
}

func prettyEncode(data interface{}, out io.Writer) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	if err := enc.Encode(data); err != nil {
		return err
	}
	return nil
}
