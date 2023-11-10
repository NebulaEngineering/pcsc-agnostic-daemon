package context

import (
	"errors"
	"strings"

	"github.com/ebfe/scard"
)

type Context struct {
	*scard.Context
}

// New create a new context to work with readers.
func New() (*Context, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, err
	}

	if ok, err := ctx.IsValid(); !ok {
		return nil, err
	}

	c := &Context{
		ctx,
	}

	return c, nil
}

// IsValid verify context
func (c *Context) IsValid() (bool, error) {
	if c.Context == nil {
		return false, errors.New("conext is not valid")
	}

	return c.Context.IsValid()

}

// Release release context
func (c *Context) Release() error {

	if c.Context != nil {
		if err := c.Context.Release(); err != nil {
			return err
		}
	}

	return nil
}

// ListReaders list readers detected in context.
func (c *Context) ListReaders() ([]string, error) {

	if c.Context == nil {
		if ctx, err := scard.EstablishContext(); err != nil {
			return nil, err
		} else {
			c.Context = ctx
		}
	}
	readers, err := c.Context.ListReaders()
	if err != nil {
		// c.Context = nil
		return nil, err
	}

	rds := make([]string, 0)
	rds = append(rds, readers...)
	return rds, nil
}

// ReaderInformation verify reader with regex "key" and return real name's reader.
func (c *Context) ReaderInformation(key string) (string, error) {

	if c.Context != nil {
		if ctx, err := scard.EstablishContext(); err != nil {
			return "", err
		} else {
			c.Context = ctx
		}
	}
	readers, err := c.Context.ListReaders()
	if err != nil {
		// c.Context = nil
		return "", err
	}

	for _, r := range readers {
		if strings.Contains(r, key) {
			return r, nil
		}
	}
	return "", nil
}
