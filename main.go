package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Alfex4936/kospell/kospell"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ─────────────────────────────────────────────────────────────
// Entry point
func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"kogrammar",
		"1.0.0",
	)

	// 1) 글자‧어절 수 세기
	countLettersTool := mcp.NewTool("count_korean_letters",
		mcp.WithDescription("Count Korean UTF-8 runes and eojeols(어절)"),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("Korean text to analyze"),
		),
	)
	s.AddTool(countLettersTool, countKoreanLettersHandler)

	// 2) 맞춤법 검사
	checkGrammarTool := mcp.NewTool("check_korean_grammar",
		mcp.WithDescription("Korean grammar / spelling checker using kospell"),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("Korean text to analyze"),
		),
	)
	s.AddTool(checkGrammarTool, checkKoreanGrammarHandler)

	// 3) 이력서 전용 통합 툴 (글자 수 + 맞춤법)
	resumeReviewTool := mcp.NewTool("resume_review",
		mcp.WithDescription("Resume-oriented review: length limit + spelling feedback"),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("Korean text to analyze"),
		),
	)
	s.AddTool(resumeReviewTool, resumeReviewHandler)

	// 4) 로마자 변환
	romanizeTool := mcp.NewTool("romanize_korean",
		mcp.WithDescription("Romanize Korean text based on the National Institute of Korean Language rules."),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("Korean text to romanize"),
		),
	)
	s.AddTool(romanizeTool, romanizeKoreanHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v\n", err)
	}
}

// ─────────────────────────────────────────────────────────────
// Handlers

func countKoreanLettersHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text, err := request.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	runes := len([]rune(text))
	words := len(strings.Fields(text))
	msg := fmt.Sprintf("총 글자 수: %d자\n총 어절 수: %d어절", runes, words)

	return mcp.NewToolResultText(msg), nil
}

func checkKoreanGrammarHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text, err := request.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	report, err := kospellReport(text)
	if err != nil {
		return nil, err // Returning a top-level error
	}

	return mcp.NewToolResultText(report), nil
}

func resumeReviewHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text, err := request.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// ① 글자 수
	runeCnt := len([]rune(text))
	charLine := fmt.Sprintf("🔢 글자 수: %d자", runeCnt)

	// ② 맞춤법
	grammarLine, err := kospellReport(text)
	if err != nil {
		return nil, err // Returning a top-level error
	}

	// ③ 종합 메시지
	var sb strings.Builder
	sb.WriteString(charLine)
	sb.WriteString("\n\n")
	sb.WriteString(grammarLine)

	return mcp.NewToolResultText(sb.String()), nil
}

func romanizeKoreanHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	text, err := request.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	romanizedText := Romanize(text)

	return mcp.NewToolResultText(romanizedText), nil
}

// ─────────────────────────────────────────────────────────────
// Helpers

// kospellReport transforms kospell.Result → human-readable summary.
func kospellReport(text string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	res, err := kospell.Check(ctx, text)
	if err != nil {
		return "", err
	}

	if res.ErrorCount == 0 {
		return "✅ 맞춤법 검사 결과: 오류가 없습니다.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("❌ 총 %d개의 오류가 발견되었습니다:\n", res.ErrorCount))

	for _, chunk := range res.Corrections {
		for _, item := range chunk.Items {
			original := item.Origin
			best := item.Suggest[0] // 첫 번째 제안을 대표로 사용
			sb.WriteString(
				fmt.Sprintf("- \"%s\" → \"%s\"\n", original, best),
			)
		}
	}
	return sb.String(), nil
}

// ─────────────────────────────────────────────────────────────
// Romanization

const (
	hangulStart = 0xAC00
	hangulEnd   = 0xD7A3
	choCount    = 19
	jungCount   = 21
	jongCount   = 28
)

var (
	choTable  = []string{"g", "kk", "n", "d", "tt", "r", "m", "b", "pp", "s", "ss", "", "j", "jj", "ch", "k", "t", "p", "h"}
	jungTable = []string{"a", "ae", "ya", "yae", "eo", "e", "yeo", "ye", "o", "wa", "wae", "oe", "yo", "u", "wo", "we", "wi", "yu", "eu", "ui", "i"}
	jongTable = []string{"", "g", "kk", "gs", "n", "nj", "nh", "d", "l", "lg", "lm", "lb", "ls", "lt", "lp", "lh", "m", "b", "bs", "s", "ss", "ng", "j", "ch", "k", "t", "p", "h"}
	// Representative sounds for final consonants when not followed by a vowel
	jongTerminalTable = []string{"", "k", "k", "k", "n", "n", "n", "t", "l", "k", "m", "p", "l", "t", "p", "t", "m", "p", "p", "t", "t", "ng", "t", "t", "k", "t", "p", "h"}
)

// hangulSyllable represents the phonetic components of a Hangul character.
type hangulSyllable struct {
	cho, jung, jong int
	isHangul        bool
	original        rune
}

// Romanize converts Korean text to Roman letters based on the official NIKL rules.
// It uses a multi-pass approach to handle complex phonetic assimilation rules.
func Romanize(text string) string {
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return ""
	}

	syllables := make([]hangulSyllable, n)

	// Pass 1: Decompose all runes into their Jamo components.
	for i, r := range runes {
		if r >= hangulStart && r <= hangulEnd {
			code := int(r - hangulStart)
			syllables[i] = hangulSyllable{
				cho:      code / (jungCount * jongCount),
				jung:     (code % (jungCount * jongCount)) / jongCount,
				jong:     code % jongCount,
				isHangul: true,
				original: r,
			}
		} else {
			syllables[i] = hangulSyllable{isHangul: false, original: r}
		}
	}

	// Pass 2: Apply sound change rules by modifying the Jamo components.
	for i := 0; i < n-1; i++ {
		s1 := &syllables[i]
		s2 := &syllables[i+1]

		if !s1.isHangul || !s2.isHangul || s1.jong == 0 {
			continue
		}

		// Rule: Palatalization (e.g., 같이 -> gachi, 해돋이 -> haedoji)
		if (s1.jong == 7 || s1.jong == 25) && s2.cho == 11 && s2.jung == 20 { // ㄷ, ㅌ + ㅣ
			s1.jong = 0 // Final consonant of s1 is removed
			s2.cho = 14 // Initial of s2 becomes ㅊ
			continue
		}

		// Rules for when the next syllable starts with a consonant
		if s2.cho != 11 { // if next is not vowel-initial 'ㅇ'
			// Rule: Nasalization (e.g., 백마 -> baengma, 신문로 -> sinmunno)
			if s2.cho == 2 || s2.cho == 6 { // next initial is ㄴ or ㅁ
				switch s1.jong {
				case 1, 2, 24: // ㄱ, ㄲ, ㅋ -> ㅇ
					s1.jong = 21
				case 7, 19, 20, 22, 23, 25: // ㄷ, ㅅ, ㅆ, ㅈ, ㅊ, ㅌ -> ㄴ
					s1.jong = 4
				case 17, 26: // ㅂ, ㅍ -> ㅁ
					s1.jong = 16
				}
			}

			// Rule: 'ㄹ' Assimilation (e.g., 신라 -> Silla, 별내 -> Byeollae)
			if s1.jong == 4 && s2.cho == 5 { // ㄴ + ㄹ -> ㄹ + ㄹ
				s1.jong = 8
				s2.cho = 5 // remains ㄹ
			} else if s1.jong == 8 && s2.cho == 2 { // ㄹ + ㄴ -> ㄹ + ㄹ
				s2.cho = 5
			}
		}
	}

	// Pass 3: Build the final Romanized string from the modified syllables.
	var sb strings.Builder
	for i := 0; i < n; i++ {
		s := syllables[i]
		if !s.isHangul {
			sb.WriteRune(s.original)
			continue
		}

		// Handle previous syllable's final for liaison
		if i > 0 {
			prev := syllables[i-1]
			if prev.isHangul && prev.jong != 0 && s.cho == 11 { // Liaison
				// A final from previous syllable moves here
				sb.WriteString(jongTable[prev.jong])
			}
		}

		sb.WriteString(choTable[s.cho])
		sb.WriteString(jungTable[s.jung])

		// Write final consonant if it doesn't cause liaison
		if s.jong != 0 {
			nextIsVowelInitial := false
			if i+1 < n && syllables[i+1].isHangul && syllables[i+1].cho == 11 {
				nextIsVowelInitial = true
			}
			if !nextIsVowelInitial {
				sb.WriteString(jongTerminalTable[s.jong])
			}
		}
	}

	return sb.String()
}
