package utils

import (
	"os"
	"path"
	"sync"

	"github.com/CircleCI-Public/circleci-yaml-language-server/pkg/ast"
	"github.com/adrg/xdg"
	"go.lsp.dev/protocol"
)

type Cache struct {
	FileCache    FileCache
	OrbCache     OrbCache
	DockerCache  DockerCache
	ContextCache ContextCache
	ProjectCache ProjectCache
}

type DockerCache struct {
	cacheMutex  *sync.Mutex
	dockerCache map[string]*CachedDockerImage
}

type CachedDockerImage struct {
	Checked bool
	Exists  bool
}

type FileCache struct {
	cacheMutex *sync.Mutex
	fileCache  map[protocol.URI]*protocol.TextDocumentItem
}

type OrbCache struct {
	cacheMutex *sync.Mutex
	orbsCache  map[string]*ast.OrbInfo
}

type ContextCache struct {
	cacheMutex   *sync.Mutex
	contextCache map[string]*Context
}

type ProjectCache struct {
	cacheMutex   *sync.Mutex
	projectCache map[string]*Project
}

func (c *Cache) init() {
	c.FileCache.fileCache = make(map[protocol.URI]*protocol.TextDocumentItem)
	c.FileCache.cacheMutex = &sync.Mutex{}

	c.OrbCache.orbsCache = make(map[string]*ast.OrbInfo)
	c.OrbCache.cacheMutex = &sync.Mutex{}

	c.DockerCache.cacheMutex = &sync.Mutex{}
	c.DockerCache.dockerCache = make(map[string]*CachedDockerImage)

	c.ContextCache.cacheMutex = &sync.Mutex{}
	c.ContextCache.contextCache = make(map[string]*Context)

	c.ProjectCache.cacheMutex = &sync.Mutex{}
	c.ProjectCache.projectCache = make(map[string]*Project)
}

// FILE

func (c *FileCache) SetFile(file *protocol.TextDocumentItem) protocol.TextDocumentItem {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.fileCache[file.URI] = file
	return *file
}

func (c *FileCache) GetFile(uri protocol.URI) *protocol.TextDocumentItem {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	return c.fileCache[uri]
}

func (c *FileCache) GetFiles() map[protocol.URI]*protocol.TextDocumentItem {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	return c.fileCache
}

func (c *FileCache) RemoveFile(uri protocol.URI) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	delete(c.fileCache, uri)
}

// ORBS

func (c *OrbCache) HasOrb(orbID string) bool {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	_, ok := c.orbsCache[orbID]

	return ok
}

func (c *OrbCache) SetOrb(orb *ast.OrbInfo, orbID string) ast.OrbInfo {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.orbsCache[orbID] = orb
	return *orb
}

func (c *OrbCache) UpdateOrbParsedAttributes(orbID string, parsedOrbAttributes ast.OrbParsedAttributes) ast.OrbParsedAttributes {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.orbsCache[orbID].OrbParsedAttributes = parsedOrbAttributes
	return parsedOrbAttributes
}

func (c *OrbCache) GetOrb(orbID string) *ast.OrbInfo {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	return c.orbsCache[orbID]
}

func (c *OrbCache) RemoveOrb(orbID string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	delete(c.orbsCache, orbID)
}

func (c *OrbCache) RemoveOrbs() {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	for k := range c.orbsCache {
		delete(c.orbsCache, k)
	}
}

func (c *Cache) RemoveOrbFiles() {
	c.OrbCache.cacheMutex.Lock()
	defer c.OrbCache.cacheMutex.Unlock()
	c.FileCache.cacheMutex.Lock()
	defer c.FileCache.cacheMutex.Unlock()

	for _, orb := range c.OrbCache.orbsCache {
		if _, err := os.Stat(orb.RemoteInfo.FilePath); err == nil {
			os.Remove(orb.RemoteInfo.FilePath)
		}
	}
}

// Docker images cache

func (c *DockerCache) Add(name string, exists bool) *CachedDockerImage {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.dockerCache[name] = &CachedDockerImage{
		Checked: true,
		Exists:  exists,
	}

	return c.dockerCache[name]
}

func (c *DockerCache) Get(name string) *CachedDockerImage {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	return c.dockerCache[name]
}

func (c *DockerCache) Remove(name string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	delete(c.dockerCache, name)
}

func CreateCache() *Cache {
	cache := Cache{}
	cache.init()
	return &cache
}

func GetOrbCacheFSPath(orbYaml string) string {
	file := path.Join("cci", "orbs", ".circleci", orbYaml+".yml")
	filePath, err := xdg.CacheFile(file)

	if err != nil {
		filePath = path.Join(xdg.Home, ".cache", file)
	}

	return filePath
}

func (cache *Cache) ClearHostData() {
	cache.RemoveOrbFiles()
	cache.OrbCache.RemoveOrbs()
}

// Context cache

func (c *ContextCache) SetContext(ctx *Context) *Context {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.contextCache[ctx.Name] = ctx
	return ctx
}

func (c *ContextCache) GetContext(name string) *Context {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	return c.contextCache[name]
}

func (c *ContextCache) RemoveContext(name string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	delete(c.contextCache, name)
}

func (c *ContextCache) AddEnvVariableToContext(name string, envVariable string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	ctx := c.contextCache[name]

	if FindInArray(ctx.envVariables, envVariable) < 0 {
		ctx.envVariables = append(ctx.envVariables, envVariable)
	}
	c.contextCache[name] = ctx
}

func (c *ContextCache) GetAllContext() map[string]*Context {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	return c.contextCache
}

// Project cache

func (c *ProjectCache) SetProject(project *Project) *Project {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	c.projectCache[project.Slug] = project
	return project
}

func (c *ProjectCache) GetProject(name string) *Project {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	return c.projectCache[name]
}

func (c *ProjectCache) RemoveProject(name string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	delete(c.projectCache, name)
}

func (c *ProjectCache) GetAllProjects() map[string]*Project {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	return c.projectCache
}

func (c *ProjectCache) AddEnvVariableToProject(name string, envVariable string) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()
	project := c.projectCache[name]

	if FindInArray(project.EnvVariables, envVariable) < 0 {
		project.EnvVariables = append(project.EnvVariables, envVariable)
	}
	c.projectCache[name] = project
}
