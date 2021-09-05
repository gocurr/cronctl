package cronctl

var (
	nameFun = make(map[string]func())
)

type Jobs []struct {
	Name string `json:"name"`
	Spec string `json:"spec"`
}

func jobinfos(jobs Jobs) []*jobinfo {
	var jobinfos []*jobinfo

	for _, job := range jobs {
		fn := fn(job.Name)
		if fn != nil {
			jobinfos = append(jobinfos, &jobinfo{
				Name: job.Name,
				Spec: job.Spec,
				Fn:   fn,
			})
		}
	}
	return jobinfos
}

func Map(name string, fn func()) {
	nameFun[name] = fn
}

func fn(name string) func() {
	nf, ok := nameFun[name]
	if !ok {
		return nil
	}
	return nf
}
