package functional

type ErrorFn = func() error

// lazy exec function on no errors
func LazyFn(fns ...ErrorFn) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
