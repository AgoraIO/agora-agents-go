package Agora

import (
	"strings"
	"testing"
)

type legacyAsrVisitor struct{}

func (legacyAsrVisitor) VisitAres(*AresAsr) error                   { return nil }
func (legacyAsrVisitor) VisitFengming(*FengmingAsr) error           { return nil }
func (legacyAsrVisitor) VisitTencent(*TencentAsr) error             { return nil }
func (legacyAsrVisitor) VisitMicrosoft(*MicrosoftAsr) error         { return nil }
func (legacyAsrVisitor) VisitDeepgram(*DeepgramAsr) error           { return nil }
func (legacyAsrVisitor) VisitOpenai(*OpenAiAsr) error               { return nil }
func (legacyAsrVisitor) VisitGoogle(*GoogleAsr) error               { return nil }
func (legacyAsrVisitor) VisitGemini(*GeminiAsr) error               { return nil }
func (legacyAsrVisitor) VisitAmazon(*AmazonAsr) error               { return nil }
func (legacyAsrVisitor) VisitAssemblyai(*AssemblyAiAsr) error       { return nil }
func (legacyAsrVisitor) VisitSpeechmatics(*SpeechmaticsAsr) error   { return nil }
func (legacyAsrVisitor) VisitSarvam(*SarvamAsr) error               { return nil }
func (legacyAsrVisitor) VisitXai(*XAiAsr) error                     { return nil }
func (legacyAsrVisitor) VisitXfyun(*XfyunAsr) error                 { return nil }
func (legacyAsrVisitor) VisitXfyunBigmodel(*XfyunBigmodelAsr) error { return nil }
func (legacyAsrVisitor) VisitXfyunDialect(*XfyunDialectAsr) error   { return nil }

type smallestAiAsrTestVisitor struct {
	legacyAsrVisitor
	visited bool
}

func (v *smallestAiAsrTestVisitor) VisitSmallestai(*SmallestAiAsr) error {
	v.visited = true
	return nil
}

type legacyTtsVisitor struct{}

func (legacyTtsVisitor) VisitTencent(*TencentTts) error                 { return nil }
func (legacyTtsVisitor) VisitBytedance(*BytedanceTts) error             { return nil }
func (legacyTtsVisitor) VisitMicrosoft(*MicrosoftTts) error             { return nil }
func (legacyTtsVisitor) VisitElevenlabs(*ElevenLabsTts) error           { return nil }
func (legacyTtsVisitor) VisitMinimax(*MinimaxTts) error                 { return nil }
func (legacyTtsVisitor) VisitMurf(*MurfTts) error                       { return nil }
func (legacyTtsVisitor) VisitCartesia(*CartesiaTts) error               { return nil }
func (legacyTtsVisitor) VisitOpenai(*OpenAiTts) error                   { return nil }
func (legacyTtsVisitor) VisitHumeai(*HumeAiTts) error                   { return nil }
func (legacyTtsVisitor) VisitRime(*RimeTts) error                       { return nil }
func (legacyTtsVisitor) VisitFishaudio(*FishAudioTts) error             { return nil }
func (legacyTtsVisitor) VisitGoogle(*GoogleTts) error                   { return nil }
func (legacyTtsVisitor) VisitAmazon(*AmazonTts) error                   { return nil }
func (legacyTtsVisitor) VisitSarvam(*SarvamTts) error                   { return nil }
func (legacyTtsVisitor) VisitGenericHTTP(*GenericHTTPTts) error         { return nil }
func (legacyTtsVisitor) VisitXai(*XAiTts) error                         { return nil }
func (legacyTtsVisitor) VisitDeepgram(*DeepgramTts) error               { return nil }
func (legacyTtsVisitor) VisitCosyvoice(*CosyvoiceTts) error             { return nil }
func (legacyTtsVisitor) VisitBytedanceDuplex(*BytedanceDuplexTts) error { return nil }
func (legacyTtsVisitor) VisitStepfun(*StepfunTts) error                 { return nil }
func (legacyTtsVisitor) VisitGradium(*GradiumTts) error                 { return nil }
func (legacyTtsVisitor) VisitMistral(*MistralTts) error                 { return nil }
func (legacyTtsVisitor) VisitTypecast(*TypecastTts) error               { return nil }

type smallestAiTtsTestVisitor struct {
	legacyTtsVisitor
	visited bool
}

func (v *smallestAiTtsTestVisitor) VisitSmallestai(*SmallestAiTts) error {
	v.visited = true
	return nil
}

var (
	_ AsrVisitor           = legacyAsrVisitor{}
	_ SmallestAiAsrVisitor = (*smallestAiAsrTestVisitor)(nil)
	_ TtsVisitor           = legacyTtsVisitor{}
	_ SmallestAiTtsVisitor = (*smallestAiTtsTestVisitor)(nil)
)

func TestAsrAcceptPreservesLegacyVisitorCompatibility(t *testing.T) {
	asr := &Asr{Smallestai: &SmallestAiAsr{}}

	err := asr.Accept(legacyAsrVisitor{})
	if err == nil || !strings.Contains(err.Error(), "does not support smallestai ASR") {
		t.Fatalf("Accept() error = %v, want unsupported smallestai ASR error", err)
	}

	visitor := &smallestAiAsrTestVisitor{}
	if err := asr.Accept(visitor); err != nil {
		t.Fatalf("Accept() error = %v", err)
	}
	if !visitor.visited {
		t.Fatal("Accept() did not dispatch to SmallestAiAsrVisitor")
	}
}

func TestTtsAcceptPreservesLegacyVisitorCompatibility(t *testing.T) {
	tts := &Tts{Smallestai: &SmallestAiTts{}}

	err := tts.Accept(legacyTtsVisitor{})
	if err == nil || !strings.Contains(err.Error(), "does not support smallestai TTS") {
		t.Fatalf("Accept() error = %v, want unsupported smallestai TTS error", err)
	}

	visitor := &smallestAiTtsTestVisitor{}
	if err := tts.Accept(visitor); err != nil {
		t.Fatalf("Accept() error = %v", err)
	}
	if !visitor.visited {
		t.Fatal("Accept() did not dispatch to SmallestAiTtsVisitor")
	}
}
