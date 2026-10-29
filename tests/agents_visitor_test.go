package main

import (
	"errors"
	"strings"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

// Embedding the base interfaces supplies unrelated methods. Optional methods
// must be implemented explicitly so each visitor supports only its own vendor.
type smallestAiAsrVisitor struct {
	Agora.AsrVisitor
	visit func(*Agora.SmallestAiAsr) error
}

func (v *smallestAiAsrVisitor) VisitSmallestai(value *Agora.SmallestAiAsr) error {
	return v.visit(value)
}

type stepfunAsrVisitor struct {
	Agora.AsrVisitor
	visit func(*Agora.StepfunAsr) error
}

func (v *stepfunAsrVisitor) VisitStepfun(value *Agora.StepfunAsr) error {
	return v.visit(value)
}

type smallestAiTtsVisitor struct {
	Agora.TtsVisitor
	visit func(*Agora.SmallestAiTts) error
}

func (v *smallestAiTtsVisitor) VisitSmallestai(value *Agora.SmallestAiTts) error {
	return v.visit(value)
}

type minimaxDuplexTtsVisitor struct {
	Agora.TtsVisitor
	visit func(*Agora.MinimaxDuplexTts) error
}

func (v *minimaxDuplexTtsVisitor) VisitMinimaxDuplex(value *Agora.MinimaxDuplexTts) error {
	return v.visit(value)
}

type sonioxTtsVisitor struct {
	Agora.TtsVisitor
	visit func(*Agora.SonioxTts) error
}

func (v *sonioxTtsVisitor) VisitSoniox(value *Agora.SonioxTts) error {
	return v.visit(value)
}

var (
	_ Agora.SmallestAiAsrVisitor    = (*smallestAiAsrVisitor)(nil)
	_ Agora.StepfunAsrVisitor       = (*stepfunAsrVisitor)(nil)
	_ Agora.SmallestAiTtsVisitor    = (*smallestAiTtsVisitor)(nil)
	_ Agora.MinimaxDuplexTtsVisitor = (*minimaxDuplexTtsVisitor)(nil)
	_ Agora.SonioxTtsVisitor        = (*sonioxTtsVisitor)(nil)
)

func TestAsrAcceptOptionalVisitors(t *testing.T) {
	t.Parallel()
	smallestai := &Agora.SmallestAiAsr{}
	stepfun := &Agora.StepfunAsr{}
	visitorErr := errors.New("asr visitor failed")
	tests := []struct {
		name       string
		asr        *Agora.Asr
		newVisitor func(error, *bool) Agora.AsrVisitor
	}{
		{
			name: "smallestai",
			asr:  &Agora.Asr{Smallestai: smallestai},
			newVisitor: func(err error, called *bool) Agora.AsrVisitor {
				return &smallestAiAsrVisitor{visit: func(value *Agora.SmallestAiAsr) error {
					*called = value == smallestai
					return err
				}}
			},
		},
		{
			name: "stepfun",
			asr:  &Agora.Asr{Stepfun: stepfun},
			newVisitor: func(err error, called *bool) Agora.AsrVisitor {
				return &stepfunAsrVisitor{visit: func(value *Agora.StepfunAsr) error {
					*called = value == stepfun
					return err
				}}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expectedErr := range []error{nil, visitorErr} {
				called := false
				visitor := tt.newVisitor(expectedErr, &called)
				if err := tt.asr.Accept(visitor); !errors.Is(err, expectedErr) {
					t.Fatalf("Accept() error = %v, want %v", err, expectedErr)
				}
				if !called {
					t.Fatal("Accept() did not pass the original value to the vendor visitor")
				}
			}

			err := tt.asr.Accept(struct{ Agora.AsrVisitor }{})
			if err == nil || !strings.Contains(err.Error(), "does not support "+tt.name+" ASR") {
				t.Fatalf("Accept() error = %v, want unsupported %s ASR visitor", err, tt.name)
			}
		})
	}
}

func TestTtsAcceptOptionalVisitors(t *testing.T) {
	t.Parallel()
	smallestai := &Agora.SmallestAiTts{}
	minimaxDuplex := &Agora.MinimaxDuplexTts{}
	soniox := &Agora.SonioxTts{}
	visitorErr := errors.New("tts visitor failed")
	tests := []struct {
		name       string
		tts        *Agora.Tts
		newVisitor func(error, *bool) Agora.TtsVisitor
	}{
		{
			name: "smallestai",
			tts:  &Agora.Tts{Smallestai: smallestai},
			newVisitor: func(err error, called *bool) Agora.TtsVisitor {
				return &smallestAiTtsVisitor{visit: func(value *Agora.SmallestAiTts) error {
					*called = value == smallestai
					return err
				}}
			},
		},
		{
			name: "minimax_duplex",
			tts:  &Agora.Tts{MinimaxDuplex: minimaxDuplex},
			newVisitor: func(err error, called *bool) Agora.TtsVisitor {
				return &minimaxDuplexTtsVisitor{visit: func(value *Agora.MinimaxDuplexTts) error {
					*called = value == minimaxDuplex
					return err
				}}
			},
		},
		{
			name: "soniox",
			tts:  &Agora.Tts{Soniox: soniox},
			newVisitor: func(err error, called *bool) Agora.TtsVisitor {
				return &sonioxTtsVisitor{visit: func(value *Agora.SonioxTts) error {
					*called = value == soniox
					return err
				}}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, expectedErr := range []error{nil, visitorErr} {
				called := false
				visitor := tt.newVisitor(expectedErr, &called)
				if err := tt.tts.Accept(visitor); !errors.Is(err, expectedErr) {
					t.Fatalf("Accept() error = %v, want %v", err, expectedErr)
				}
				if !called {
					t.Fatal("Accept() did not pass the original value to the vendor visitor")
				}
			}

			err := tt.tts.Accept(struct{ Agora.TtsVisitor }{})
			if err == nil || !strings.Contains(err.Error(), "does not support "+tt.name+" TTS") {
				t.Fatalf("Accept() error = %v, want unsupported %s TTS visitor", err, tt.name)
			}
		})
	}
}
