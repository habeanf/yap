package eval

// import "log"

func Precision(truePositives, testPositives int) float64 {
	return float64(truePositives) / float64(testPositives)
}

func Recall(truePositives, conditionPositives int) float64 {
	return float64(truePositives) / float64(conditionPositives)
}

func F1(precision, recall float64) float64 {
	return 2.0 * (precision * recall) / (precision + recall)
}

type Error interface {
	String() string
	Class() string
}

type Errors []Error

func (ers Errors) ByType() map[string]int {
	retval := make(map[string]int)
	for _, e := range ers {
		if curval, exists := retval[e.Class()]; exists {
			retval[e.Class()] = curval + 1
		} else {
			retval[e.Class()] = 1
		}
	}
	return retval
}

type Result struct {
	TP, FP, TN, FN int
	Errors         Errors
	Other          interface{}
}

func (r *Result) All() int {
	return r.TP + r.FP + r.TN + r.FN
}

func (r *Result) Correct() int {
	return r.TP + r.TN
}

func (r *Result) Incorrect() int {
	return r.FP + r.FN
}

func (r *Result) TestPositives() int {
	return r.TP + r.FP
}

func (r *Result) TestNegatives() int {
	return r.TN + r.FN
}

func (r *Result) ConditionPositives() int {
	return r.TP + r.TN
}

func (r *Result) ConditionNegatives() int {
	return r.FP + r.FN
}

func (r *Result) Precision() float64 {
	return Precision(r.TP, r.TestPositives())
}

func (r *Result) Recall() float64 {
	return Recall(r.TP, r.ConditionPositives())
}

func (r *Result) Accuracy() float64 {
	return float64(r.Correct()) / float64(r.All())
}

func (r *Result) F1() float64 {
	// log.Println("Calculating F1 for Precision, Recall of", r.Precision(), r.Recall())
	return F1(r.Precision(), r.Recall())
}

type Eval func(test, condition interface{}) *Result

type Total struct {
	Result
	Results           []*Result
	Exact, Population int
}

func (t *Total) Add(r *Result) {
	t.TP += r.TP
	t.FP += r.FP
	t.TN += r.TN
	t.FN += r.FN
	if r.Incorrect() == 0 {
		t.Exact += 1
	}
	t.Population += 1
	if t.Results != nil {
		t.Results = append(t.Results, r)
	}
}

func (t *Total) ExactMatch() float64 {
	return float64(t.Exact) / float64(t.Population)
}

func (t *Total) Errors() Errors {
	retval := make([]Error, t.Incorrect())
	for _, v := range t.Results {
		if v.Errors != nil {
			retval = append(retval, v.Errors...)
		}
	}
	return retval
}
