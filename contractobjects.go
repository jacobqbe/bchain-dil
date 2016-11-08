package main

type Policy struct {
	ID string `json:"id"`
	HolderID string `json:"holderID"`
	Countries []string `json:"countries"`
	Terms []CarrierTerms `json:"terms"`
	Votes []Approval `json:"votes"`
}

type AllPolicies struct {
	Catalog []Policy `json:"policies"`
}

type Approval struct {
	CarrierID string `json:"carrier"`
	Vote string `json:"vote"`
}

type PolicyHolder struct {
	ID string `json:"id"`
	Policies []Policy `json:"policies"`
}

type AllHolders struct {
	Catalog []PolicyHolder
}

type CarrierTerms struct {
	CarrierID string `json:"carrier"`
	ID string `json:"id"`
	Country string `json:"country"`
	Premium int64 `json:"premium"`
	Value int64 `json:"value"`
}
