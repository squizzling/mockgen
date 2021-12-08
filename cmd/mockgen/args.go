package main

type Opts struct {
	Module  string   `short:"m" long:"module"`
	Package []string `short:"p" long:"package"`
	File    string   `short:"f" long:"file" required:"true"`
}

func (o *Opts) Validate() []string {
	return nil
}
