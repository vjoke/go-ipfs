package commands

import (
	"io"
)

type ChannelMarshaler struct {
	Channel   <-chan interface{}
	Marshaler func(interface{}) (io.Reader, error)
	Res       Response

	reader io.Reader
}

func (cr *ChannelMarshaler) Read(p []byte) (int, error) {
	for {
		if cr.reader == nil {
			val, more := <-cr.Channel
			if !more {
				//check error in response
				if cr.Res.Error() != nil {
					return 0, cr.Res.Error()
				}
				return 0, io.EOF
			}

			r, err := cr.Marshaler(val)
			if err != nil {
				return 0, err
			}
			if r == nil {
				continue
			}
			cr.reader = r
		}

		n, err := cr.reader.Read(p)
		if err == io.EOF {
			cr.reader = nil
			err = nil
			if n == 0 {
				continue
			}
		}
		return n, err
	}
}
