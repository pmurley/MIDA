package fetch

import b "github.com/pmurley/mida/base"

func FromFile(fileName string) (<-chan b.RawTask, error) {
	taskSet, err := b.ReadTasksFromFile(fileName)
	if err != nil {
		return nil, err
	}

	res := make(chan b.RawTask)

	go func() {
		for _, task := range taskSet {
			res <- task
		}
		close(res)
	}()

	return res, nil
}
