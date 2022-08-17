package mos

import (
	"fmt"

	"github.com/kudrykv/latex-yearly-planner/app2/contents"
	"github.com/kudrykv/latex-yearly-planner/app2/pages2"
	"github.com/kudrykv/latex-yearly-planner/app2/planners/common"
	"github.com/kudrykv/latex-yearly-planner/app2/types"
	"github.com/kudrykv/latex-yearly-planner/lib/texcalendar"
)

type Planner struct {
	layout     types.Layout
	parameters Parameters
}

func New(layout types.Layout) (*Planner, error) {
	parameters, ok := layout.Misc.(Parameters)
	if !ok {
		return nil, fmt.Errorf("expected Parameters, got %T", layout.Misc)
	}

	if err := parameters.test(); err != nil {
		return nil, fmt.Errorf("test parameters: %w", err)
	}

	parameters.calendar = texcalendar.NewYear(parameters.Year)

	planner := &Planner{layout: layout, parameters: parameters}
	if err := planner.test(); err != nil {
		return nil, fmt.Errorf("test planner: %w", err)
	}

	return planner, nil
}

func (r *Planner) test() error {
	for section := range r.sections() {
		if !Contains(r.parameters.EnabledSections, section) {
			return fmt.Errorf("unknown section %s", section)
		}
	}

	return nil
}

func (r *Planner) BuildData() (types.NamedDatas, error) {
	sections := r.sections()

	result := make(types.NamedDatas, 0, len(r.parameters.EnabledSections))

	for _, name := range r.parameters.EnabledSections {
		section, ok := sections[name]
		if !ok {
			panic(fmt.Sprintf("unknown section %s", name))
		}

		buff, err := section()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}

		result = append(result, types.NamedData{Name: name, Data: buff})
	}

	return result, nil
}

type sectionFunc func() ([]byte, error)

func (r *Planner) sections() map[string]sectionFunc {
	return map[string]sectionFunc{
		common.TitleSection:   r.titleSection,
		common.DailiesSection: r.dailiesSection,
	}
}

func (r *Planner) RunTimes() int {
	return 2
}

func Contains[T comparable](slice []T, item T) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}

	return false
}

func (r *Planner) titleSection() ([]byte, error) {
	buffer := pages2.NewBuffer()

	if err := buffer.WriteBlocks(contents.NewTitle(r.parameters.calendar.Name())); err != nil {
		return nil, fmt.Errorf("write title: %w", err)
	}

	return buffer.Bytes(), nil
}

func (r *Planner) dailiesSection() ([]byte, error) {
	buffer := pages2.NewBuffer()

	for _, day := range r.parameters.calendar.Days() {
		daily := contents.NewDaily(day, r.parameters.DailyParameters)

		if err := buffer.WriteBlocks(daily); err != nil {
			return nil, fmt.Errorf("write daily: %w", err)
		}
	}

	return buffer.Bytes(), nil
}