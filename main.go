package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
	mcp_stdio "github.com/metoro-io/mcp-golang/transport/stdio"

	"github.com/Alfex4936/kospell/kospell"
)

// ─────────────────────────────────────────────────────────────
// Request payload shared by all tools
type TextRequest struct {
	Text string `json:"text" jsonschema:"required,description=Korean text to analyze"`
}

// ─────────────────────────────────────────────────────────────
// Entry point
func main() {
	done := make(chan struct{}) // 이렇게 해야 long running

	server := mcp.NewServer(mcp_stdio.NewStdioServerTransport())

	// 1) 글자‧어절 수 세기
	must(server.RegisterTool(
		"count_korean_letters",
		"Count Korean UTF-8 runes and eojeols(어절)",
		func(req TextRequest) (*mcp.ToolResponse, error) {
			runes := len([]rune(req.Text))
			words := len(strings.Fields(req.Text))
			msg := fmt.Sprintf("총 글자 수: %d자\n총 어절 수: %d어절", runes, words)
			return mcp.NewToolResponse(mcp.NewTextContent(msg)), nil
		},
	))

	// 2) 맞춤법 검사
	must(server.RegisterTool(
		"check_korean_grammar",
		"Korean grammar / spelling checker using kospell",
		func(req TextRequest) (*mcp.ToolResponse, error) {
			report, err := kospellReport(req.Text)
			if err != nil {
				return nil, err
			}
			return mcp.NewToolResponse(mcp.NewTextContent(report)), nil
		},
	))

	// 3) 이력서 전용 통합 툴 (글자 수 + 맞춤법)
	must(server.RegisterTool(
		"resume_review",
		"Resume-oriented review: length limit + spelling feedback",
		func(req TextRequest) (*mcp.ToolResponse, error) {

			// ① 글자 수
			runeCnt := len([]rune(req.Text))
			charLine := fmt.Sprintf("🔢 글자 수: %d자", runeCnt)

			// ② 맞춤법
			grammarLine, err := kospellReport(req.Text)
			if err != nil {
				return nil, err
			}

			// ③ 종합 메시지
			var sb strings.Builder
			sb.WriteString(charLine)
			sb.WriteString("\n\n")
			sb.WriteString(grammarLine)

			return mcp.NewToolResponse(mcp.NewTextContent(sb.String())), nil
		},
	))

	// Start MCP server (stdio)
	must(server.Serve())

	<-done
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

// must is a tiny helper to crash fast on init-time errors.
func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
