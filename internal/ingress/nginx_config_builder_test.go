// Copyright 2025 The Zeus Authors.
// Licensed under the Apache License 2.0. See the LICENSE file for details.

package ingress

import "testing"


func TestConfigBuilderNewLine(t *testing.T) {
  w := NewConfigBuilder()

  w.writeln("1)", "a.", "b")
  w.writeln("2)", "c.", "d")

  expected := "1)a.b\n2)c.d\n"
  got := string(w.content())
  if got != expected {
    t.Errorf("writing multiple new lines faild; got '%q', expected '%q'", got, expected)
  }

  w.writeln("3)", "-", "-")

  expected = "1)a.b\n2)c.d\n3)--\n"
  got = string(w.content())
  if got != expected {
    t.Errorf("writing multiple new lines faild; got '%q', expected '%q'", got, expected)
  }
}

func TestConfigBuilderIndent(t *testing.T) {
  w := NewConfigBuilder()

  w.intend()
  w.writeln("1)", "a.", "b")

  expected := "\t1)a.b\n"
  got := string(w.content())
  if got != expected {
    t.Errorf("writing multiple new lines faild; got '%q', expected '%q'", got, expected)
  }

  w.intend()
  w.intend()

  got = string(w.content())
  if got != expected {
    t.Errorf("writing with multiple indent faild; got '%q', expected '%q'", got, expected)
  }

  w.unintend()
  w.unintend()
  got = string(w.content())
  if got != expected {
    t.Errorf("writing with multiple indent faild; got '%q', expected '%q'", got, expected)
  }

  w.intend()
  w.writeln("1)", "a.", "b")
  got = string(w.content())
  expected = "\t1)a.b\n\t\t1)a.b\n"
  if got != expected {
    t.Errorf("writing with multiple indent faild; got '%q', expected '%q'", got, expected)
  }
}
