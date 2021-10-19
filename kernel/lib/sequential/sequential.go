package parallel

type Task func() error

func Execute(tasks []Task) error {
	for _, task := range tasks {
		if err := task(); err != nil {
			return err
		}
	}
	return nil
}
