package main

import (
	"log"
	"net/url"
	"testing"
)

func TestParsePath(t *testing.T) {
	t.Run("/hoges/abc1?q=2&foo=bar", func(t *testing.T) {
		p, err := parsePath("/hoges/abc1?q=2&foo=bar")
		if err != nil {
			log.Fatal(err)
		}
		if p.rName != "hoges" {
			t.Errorf("resource name expected %s, got %s", "hoges", p.rName)
		}
		if p.id != "abc1" {
			t.Errorf("id expected %s, got %s", "abc1", p.id)
		}
		if p.params.Get("q") != "2" || p.params.Get("foo") != "bar" {
			t.Errorf("params expected %s, got %s", url.Values{"q": []string{"2"}, "foo": []string{"bar"}}, *(p.params))
		}
	})

	t.Run("/hoges?q=2&foo=bar (no id)", func(t *testing.T) {
		p, err := parsePath("/hoges?q=2&foo=bar")
		if err != nil {
			log.Fatal(err)
		}
		if p.rName != "hoges" {
			t.Errorf("resource name expected %s, got %s", "hoges", p.rName)
		}
		if p.id != "" {
			t.Errorf("id expected %s, got %s", "", p.id)
		}
		if p.params.Get("q") != "2" || p.params.Get("foo") != "bar" {
			t.Errorf("params expected %s, got %s", url.Values{"q": []string{"2"}, "foo": []string{"bar"}}, *(p.params))
		}
	})

	t.Run("/hoges/css/style.css (no params)", func(t *testing.T) {
		p, err := parsePath("/hoges/css/style.css")
		if err != nil {
			log.Fatal(err)
		}
		if p.rName != "hoges" {
			t.Errorf("resource name expected %s, got %s", "hoges", p.rName)
		}
		if p.id != "css/style.css" {
			t.Errorf("id expected %s, got %s", "abc1", p.id)
		}
		if p.params != nil {
			t.Errorf("params should be nil, got %s", *(p.params))
		}
	})
}
