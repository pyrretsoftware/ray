package main

type rerrType struct {
	Notify func(message string, err error)
	Fatal func(message string, err error)
}

var rerr = rerrType{
	Notify: func(message string, err error) {
		if err != nil {
			rlog.Notify(message, "err")
		}
	},
	Fatal: func (message string, err error)  {
		if err != nil {
			rlog.Fatal(message)
		}
	},
}