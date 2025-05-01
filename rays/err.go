package main

type rerrType struct {
	Notify func(message string, err error, withMessage ...bool)
	Fatal func(message string, err error, withMessage ...bool)
}

var rerr = rerrType{
	Notify: func(message string, err error, withMessage ...bool) {
		if err != nil {
			if len(withMessage) > 0 && withMessage[0] {
				message += err.Error()
			}
			rlog.Notify(message, "err")
		}
	},
	Fatal: func (message string, err error, withMessage ...bool)  {
		if err != nil {
			if len(withMessage) > 0 && withMessage[0] {
				message += err.Error()
			}
			rlog.Fatal(message)
		}
	},
}