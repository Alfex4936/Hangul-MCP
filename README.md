# Hangul-MCP (한국어 유틸리티)

Hangul-MCP는 MCP(Model Chaining Protocol) 기반 한국어 유틸리티 서비스입니다. [kospell](https://github.com/Alfex4936/kospell) 라이브러리를 사용하여 한국어 맞춤법 검사, 로마자 변환 등의 기능을 제공합니다.

<p align="center">
  <img src="https://github.com/user-attachments/assets/b95384dd-b046-4be5-9abe-92e767e46464" width="80%"/>
  <br/>
  <img src="https://github.com/user-attachments/assets/36a4ddd5-e4a4-4262-a909-15c971165c37" width="80%"/>
  <br/>
  <img src="https://github.com/user-attachments/assets/952f82f4-c3a2-445f-bc6d-ac7f383b646b" width="80%"/>
</p>

## 기능

KoGrammar는 다음과 같은 기능을 제공합니다:

1. **맞춤법 검사**: 한국어 문장의 맞춤법을 검사하고 오류를 수정 제안합니다.
2. **글자 수 세기**: 한국어 텍스트의 글자 수와 어절 수를 계산합니다.
3. **이력서 리뷰**: 이력서 작성에 특화된 기능으로, 글자 수 제한 확인과 맞춤법 피드백을 함께 제공합니다.
4. **로마자 변환**: 한국어 텍스트를 국립국어원 로마자 표기법에 따라 변환합니다.

## MCP 도구 목록

KoGrammar는 다음과 같은 MCP 도구를 제공합니다:

1. `count_korean_letters`: 한국어 텍스트의 글자 수와 어절 수를 계산합니다.
2. `check_korean_grammar`: 한국어 문장의 맞춤법을 검사하고 수정 제안을 제공합니다.
3. `resume_review`: 이력서에 특화된 리뷰 기능으로, 글자 수 확인과 맞춤법 검사를 통합 제공합니다.
4. `romanize_korean`: 국립국어원 표기법에 따라 한국어 텍스트를 로마자로 변환합니다.

## 설치 방법

### 바이너리 다운로드

Windows 또는 Linux 환경에서 실행할 수 있는 바이너리를 제공합니다:
- Windows: `kogrammar.exe`
- Linux: `kogrammar-linux-amd64`

### 소스코드 빌드

소스코드에서 직접 만들려면 다음 단계를 따르세요:

1. Go 1.24 이상 설치
2. 저장소 클론
   ```
   git clone https://github.com/Alfex4936/Hangul-MCP
   cd Hangul-MCP
   ```
3. 의존성 설치
   ```
   go mod download
   ```
4. 빌드
   ```
   go build -o kogrammar
   ```

## 사용 방법

KoGrammar는 MCP 프로토콜을 통해 다른 애플리케이션과 통합되어 사용될 수 있습니다.

`settings.json` 또는 이에 준하는 환경 설정 파일에 다음과 같이 서버를 등록할 수 있습니다.

```json
{
  "mcpServers": {
    "kogrammar-server": {
      "command": "path/to/your/kogrammar.exe",
      "args": [],
      "env": {}
    }
  }
}
```

## Credits

- (주)나라인포테크의 맞춤법 검사기
- [kospell](https://github.com/Alfex4936/kospell) - 한국어 맞춤법 검사 라이브러리
- [mcp-go](https://github.com/mark3labs/mcp-go) - Model Chaining Protocol Go implementation