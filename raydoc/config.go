package main

type Job struct {
	Name string
	Path string
	Base string
	Output string
}

type raydocConfig struct {
	Jobs []Job
}