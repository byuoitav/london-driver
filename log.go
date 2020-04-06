package london

// Logger is a logger usable by DSP.
type Logger interface {
	Debugf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

func (d *DSP) debugf(format string, a ...interface{}) {
	if d.logger != nil {
		d.logger.Debugf(format, a...)
	}
}

func (d *DSP) infof(format string, a ...interface{}) {
	if d.logger != nil {
		d.logger.Infof(format, a...)
	}
}

func (d *DSP) warnf(format string, a ...interface{}) {
	if d.logger != nil {
		d.logger.Warnf(format, a...)
	}
}

func (d *DSP) errorf(format string, a ...interface{}) {
	if d.logger != nil {
		d.logger.Errorf(format, a...)
	}
}
