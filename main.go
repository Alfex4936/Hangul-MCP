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
