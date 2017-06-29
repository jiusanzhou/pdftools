package pdf2html

import (
	"fmt"
	"github.com/jiusanzhou/pdf2html/pkg/util"
	"os"
	"path"
	"path/filepath"
	"time"
)

// material of factory input
type Material struct {
	// TODO: support http get file
	Url string

	// file path
	FilePath string

	// output file path
	OutputFilePath string
}

// product of factory output
type Product struct {
	// status of this convert
	// 0: normal
	// 1: error
	// ...
	Status int

	// file path of output file
	FilePath string

	// size of out file
	Size int64

	// coast time of born
	Coast time.Duration

	// Material
	Material *Material
}

// factory for transforming pdf to html
type Factory struct {
	// configuration
	config *Config

	// execute path of pdf2htmlEx
	cmdTpl string

	// material channel for put
	in chan *Material

	// product channel for get
	out chan *Product
}

var (
	defaultData = ".data"

	defaultPdf2htmlDir     = "pdf2htmlEx"
	defaultPdf2htmlDataDir = "data"

	defaultExec = path.Join(defaultData, defaultPdf2htmlDir, "pdf2htmlEx")

	defaultExecTpl = "{{exe}} --data-dir={{data}} {{input}} {{output}}"
)

func NewFactory(c *Config) (f *Factory, err error) {

	// exec := c.exec
	// execTpl := c.execTpl

	var exec, exDataDir, execTpl string

	if c.Exec != "" {
		exec = c.Exec
	} else {
		exec = defaultExec
	}

	if c.Pdf2htmlDataDir != "" {
		exDataDir = c.Pdf2htmlDataDir
	} else {
		exDataDir = path.Join(filepath.Dir(exec), defaultPdf2htmlDataDir)
	}

	if c.ExecTpl != "" {
		execTpl = c.ExecTpl
	} else {
		execTpl = defaultExecTpl
	}

	if c.OutputDir != "" {
		// TODO: make sure it writable
	}

	f = &Factory{
		config: c,
		cmdTpl: util.ExecTpl(execTpl, map[string]string{"exe": exec, "data": exDataDir}),

		in:  make(chan *Material),
		out: make(chan *Product),
	}

	go f.Start()

	return
}

func (f *Factory) NewMaterial(filePath, outputDir, outputFileName string) (m *Material, err error) {

	// TODO: check if file path is url

	// TODO: check if file exits and is legal pdf file

	// TODO: check if output directory is writable

	var name, outputPath string
	if outputFileName == "" {
		_, name = path.Split(filePath)
	} else {
		name = outputFileName
	}

	// replace suffix
	name = name[:len(name)-len(path.Ext(name))] + ".html"

	if outputDir != "" {
		outputPath = path.Join(outputDir, name)
	} else {
		outputPath = path.Join(f.config.OutputDir, name)
	}

	m = &Material{
		FilePath:       filePath,
		OutputFilePath: outputPath,
	}

	return
}

func (f *Factory) Convert(m *Material) (p *Product, err error) {

	// var outFileName string

	var output string

	if filepath.IsAbs(m.OutputFilePath) {
		//  is abs path
		base := filepath.Dir(m.OutputFilePath)
		_, name := filepath.Split(m.OutputFilePath)
		output = "--dest-dir=" + base + " " + name

	} else {
		output = m.OutputFilePath
	}

	cmd := util.ExecTpl(f.cmdTpl, map[string]string{"input": m.FilePath, "output": output})

	p = &Product{
		Material: m,
	}

	startTime := time.Now()
	err = util.DoCmd(cmd)
	coast := time.Now().Sub(startTime)

	p.Coast = coast

	if err != nil {
		fmt.Println("转换PDF有错误,", err.Error())
		p.Status = 1
		return
	}

	fi, err := os.Stat(m.OutputFilePath)
	if err != nil {
		fmt.Println("PDF->HTML输出的文件,", err.Error())
		p.Status = 1
		return
	}

	p.Status = 0
	p.FilePath = m.OutputFilePath
	p.Size = fi.Size()
	return
}

func (f *Factory) Put(m *Material) (err error) {

	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	// log

	f.in <- m

	return
}

func (f *Factory) Get() (p *Product, err error) {
	defer func() {
		if err := recover(); err != nil {

		}
	}()

	return <-f.out, nil
}

func (f *Factory) Start() {
	defer func() {
		if err := recover(); err != nil {

		}
	}()

	for m := range f.in {

		// get a material

		// convert
		go func(m *Material) {
			p, _ := f.Convert(m)
			// put product
			f.out <- p
		}(m)

		// log error
	}
}

func (f *Factory) Close() {
	// TODO: wait finished all

	defer func() {
		if err := recover(); err != nil {
		}
	}()
	close(f.in)
	close(f.out)
}
