package main_test

import (
	"errors"
	"strings"
	"testing"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
)

type stepfunOnlyAsrVisitor struct{ Agora.AsrVisitor }

func (stepfunOnlyAsrVisitor) VisitStepfun(*Agora.StepfunAsr) error { return nil }

type elevenLabsOnlyAsrVisitor struct{ Agora.AsrVisitor }

func (elevenLabsOnlyAsrVisitor) VisitElevenlabs(*Agora.ElevenLabsAsr) error { return nil }

type fishAudioOnlyAsrVisitor struct{ Agora.AsrVisitor }

func (fishAudioOnlyAsrVisitor) VisitFishaudio(*Agora.FishAudioAsr) error { return nil }

type qwenOnlyAsrVisitor struct{ Agora.AsrVisitor }

func (qwenOnlyAsrVisitor) VisitQwen(*Agora.QwenAsr) error { return nil }

func TestAsrAcceptOptionalVisitors(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		asr     Agora.Asr
		visitor Agora.AsrVisitor
	}{
		{"stepfun", Agora.Asr{Stepfun: &Agora.StepfunAsr{}}, stepfunOnlyAsrVisitor{}},
		{"elevenlabs", Agora.Asr{Elevenlabs: &Agora.ElevenLabsAsr{}}, elevenLabsOnlyAsrVisitor{}},
		{"fishaudio", Agora.Asr{Fishaudio: &Agora.FishAudioAsr{}}, fishAudioOnlyAsrVisitor{}},
		{"qwen", Agora.Asr{Qwen: &Agora.QwenAsr{}}, qwenOnlyAsrVisitor{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.asr.Accept(tt.visitor); err != nil {
				t.Fatalf("Accept() with only the matching optional visitor: %v", err)
			}
			// A visitor for a different vendor must return an error without panicking.
			var unsupported Agora.AsrVisitor = stepfunOnlyAsrVisitor{}
			if tt.name == "stepfun" {
				unsupported = qwenOnlyAsrVisitor{}
			}
			if err := tt.asr.Accept(unsupported); err == nil || !strings.Contains(err.Error(), "does not support "+tt.name+" ASR") {
				t.Fatalf("Accept() with an unsupported visitor = %v", err)
			}
		})
	}
}

type qwenAsrCallbackVisitor struct {
	Agora.AsrVisitor
	visit func(*Agora.QwenAsr) error
}

func (v qwenAsrCallbackVisitor) VisitQwen(asr *Agora.QwenAsr) error { return v.visit(asr) }

func TestAsrAcceptPassesPayloadAndError(t *testing.T) {
	t.Parallel()
	asr := &Agora.QwenAsr{}
	wantErr := errors.New("visitor error")
	called := false
	visitor := qwenAsrCallbackVisitor{visit: func(got *Agora.QwenAsr) error {
		called = true
		if got != asr {
			t.Fatal("Accept() passed a different ASR configuration")
		}
		return wantErr
	}}
	if err := (&Agora.Asr{Qwen: asr}).Accept(visitor); !errors.Is(err, wantErr) {
		t.Fatalf("Accept() = %v, want %v", err, wantErr)
	}
	if !called {
		t.Fatal("Accept() did not call the Qwen visitor")
	}
}
