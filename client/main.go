// file: kogrammar_client.go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

// 작은 헬퍼
func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// 1) 서버 프로세스 실행
	cmd := exec.Command("D:/Dev/mcp/kogrammar/kogrammar.exe")

	stdin, err := cmd.StdinPipe()
	must(err)
	stdout, err := cmd.StdoutPipe()
	must(err)

	must(cmd.Start())
	defer cmd.Process.Kill() // 프로그램 종료 시 서버도 종료

	// 2) stdio 클라이언트 트랜스포트 구성
	transport := stdio.NewStdioServerTransportWithIO(stdout, stdin)
	client := mcp.NewClient(transport)

	// 3) MCP 초기화
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := client.Initialize(ctx); err != nil {
		log.Fatalf("initialize 실패: %v", err)
	}

	// 4) 테스트용 입력
	text := "안녕하세요, 저는언제나 활기찬 바라보게 하겠읍니다."
	args := map[string]interface{}{"text": text}

	tools := []string{
		"count_korean_letters",
		"check_korean_grammar",
		"resume_review",
	}

	for _, tool := range tools {
		resp, err := client.CallTool(ctx, tool, args)
		if err != nil {
			log.Printf("[%s] 호출 실패: %v", tool, err)
			continue
		}
		if len(resp.Content) > 0 && resp.Content[0].TextContent != nil {
			fmt.Printf("\n── %s 결과 ──\n%s\n", tool, resp.Content[0].TextContent.Text)
		}
	}

	// 5) stdin 닫아서 서버 graceful-exit 유도
	stdin.Close()
	io.Copy(io.Discard, stdout)
	cmd.Wait()
}
