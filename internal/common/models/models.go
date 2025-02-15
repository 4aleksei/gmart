package models

import (
	"encoding/json"
	"io"
	"time"
	//"strconv"
)

type (
	UserRegistration struct {
		Name     string `json:"login"`
		Password string `json:"password"`
	}

	Order struct {
		OrderID string    `json:"number"`
		Status  string    `json:"status"`
		Accrual float64   `json:"accrual,omitempty"`
		Time    time.Time `json:"uploaded_at"`
	}

	Balance struct {
		Accrual   float64 `json:"current"`
		Withdrawn uint32  `json:"withdrawn"`
	}

	Withdraw struct {
		OrderID string    `json:"order"`
		Sum     float64   `json:"sum"`
		TimeC   time.Time `json:"processed_at,omitempty"`
	}
)

func (val *Withdraw) FromJSON(body io.ReadCloser) error {
	err := json.NewDecoder(body).Decode(val)
	return err
}

func (val *UserRegistration) FromJSON(body io.ReadCloser) error {
	err := json.NewDecoder(body).Decode(val)
	return err
}
func (val *UserRegistration) ToJSON(w io.Writer) error {
	err := json.NewEncoder(w).Encode(val)
	return err
}

func JSONSEncodeBytes(w io.Writer, val any /*val []Order*/) error {
	enc := json.NewEncoder(w)
	err := enc.Encode(val)
	return err
}

/*
func (valModels *Metrics) ConvertMetricToModel(name string, valMetrics valuemetric.ValueMetric) {
	valModels.ID = name
	valModels.MType = valMetrics.GetTypeStr()
	valModels.Delta = valMetrics.ValueInt()
	valModels.Value = valMetrics.ValueFloat()
}

func (valModels *Metrics) ConvertMetricToValue() string {
	if valModels.Delta != nil {
		return strconv.FormatInt(*valModels.Delta, 10)
	} else if valModels.Value != nil {
		return strconv.FormatFloat(*valModels.Value, 'f', -1, 64)
	}
	return ""
}

func (valModels *Metrics) JSONDecode(body io.ReadCloser) error {
	dec := json.NewDecoder(body)
	err := dec.Decode(valModels)
	return err
}

func (valModels *Metrics) JSONEncodeBytes(w io.Writer) error {
	enc := json.NewEncoder(w)
	err := enc.Encode(valModels)
	return err
}

func JSONSDecode(body io.ReadCloser) ([]Metrics, error) {
	var valModels []Metrics
	dec := json.NewDecoder(body)
	err := dec.Decode(&valModels)
	return valModels, err
}

func JSONSEncodeBytes(w io.Writer, val []Metrics) error {
	enc := json.NewEncoder(w)
	err := enc.Encode(val)
	return err
}
*/
