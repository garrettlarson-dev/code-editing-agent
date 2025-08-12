package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/anthropics/anthropic-sdk-go"
)

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	tools          []ToolDefinition
}

type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

func main() {
	client := anthropic.NewClient()

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	tools := []ToolDefinition{ReadFileDefinition, ListFilesDefinition, EditFileDefinition}
	agent := NewAgent(&client, getUserMessage, tools)

	err := agent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func NewAgent(
	client *anthropic.Client,
	getUserMessage func() (string, bool),
	tools []ToolDefinition,
) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			case "tool_use":
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
		conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
	}

	return nil
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDef ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	if !found {
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}
	return anthropic.NewToolResultBlock(id, response, false)
}	

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}
	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(1024),
		Messages:  conversation,
		Tools:     anthropicTools,
	})
	return message, err
}

var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		panic(err)
	}

	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}


var ListFilesDefinition = ToolDefinition{
	Name:        "list_files",
	Description: "List files and directories at a given path. If no path is provided, lists files in the current directory.",
	InputSchema: ListFilesInputSchema,
	Function:    ListFiles,
}

type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

var ListFilesInputSchema = GenerateSchema[ListFilesInput]()

func ListFiles(input json.RawMessage) (string, error) {
	listFilesInput := ListFilesInput{}
	err := json.Unmarshal(input, &listFilesInput)
	if err != nil {
		panic(err)
	}

	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath != "." {
			if info.IsDir() {
				files = append(files, relPath+"/")
			} else {
				files = append(files, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

type EditMode string

const (
  ModeReplaceOnce EditMode = "replace_once"  // replace exactly N matches (default 1)
  ModeWriteFull   EditMode = "write_full"    // replace entire file with New
  ModeAppend      EditMode = "append"        // append New to EOF (with optional newline)
)

type EditFileInput struct {
  Path        string   `json:"path" jsonschema_description:"Relative path to the file (inside the working directory)"`
  Mode        EditMode `json:"mode" jsonschema_description:"One of: replace_once | write_full | append"`
  OldStr      string   `json:"old_str,omitempty" jsonschema_description:"Exact text to replace (required for replace_once)"`
  NewStr      string   `json:"new_str,omitempty" jsonschema_description:"Text to write or replace with"`
  ExpectCount int      `json:"expect_count,omitempty" jsonschema_description:"Expected number of matches for replace_once (default 1)"`
  ExpectHash  string   `json:"expect_hash,omitempty" jsonschema_description:"Optional SHA-256 of the current file for idempotency; if provided and mismatched, the tool will fail"`
  EnsureNL    bool     `json:"ensure_trailing_newline,omitempty" jsonschema_description:"If true and mode=append, ensures a newline before appending"`
}

type EditResult struct {
  Changed       bool   `json:"changed"`
  OldHash       string `json:"old_hash"`
  NewHash       string `json:"new_hash"`
  Replacements  int    `json:"replacements"`
  Message       string `json:"message"`
}

var EditFileDefinition = ToolDefinition{
  Name: "edit_file",
  Description: `Edit a text file deterministically.
- mode=replace_once: replace OldStr with NewStr exactly ExpectCount times (default 1). If match count differs, the tool fails (no changes).
- mode=write_full: write NewStr as the entire file (use for creating or full rewrites).
- mode=append: append NewStr to the end of the file. If EnsureNL=true, add a newline if missing.
- If ExpectHash is set and does not match the file's current hash, the tool fails (no changes). 
Return JSON with Changed, hashes, and Replacements.
If no change would occur, succeed with Changed=false (do NOT loop).`,
  InputSchema: EditFileInputSchema,
  Function:    EditFile,
}

var EditFileInputSchema = GenerateSchema[EditFileInput]()

func fileHash(b []byte) string {
  h := sha256.Sum256(b)
  return hex.EncodeToString(h[:])
}

func safePath(p string) (string, error) {
  p = filepath.Clean(p)
  if strings.HasPrefix(p, "..") || filepath.IsAbs(p) {
    return "", errors.New("path must be inside working directory")
  }
  return p, nil
}

func EditFile(input json.RawMessage) (string, error) {
  var in EditFileInput
  if err := json.Unmarshal(input, &in); err != nil {
    return "", err
  }
  if in.Path == "" || in.Mode == "" {
    return "", fmt.Errorf("path and mode are required")
  }
  if in.Mode == ModeReplaceOnce && in.OldStr == "" {
    return "", fmt.Errorf("old_str required for mode=replace_once")
  }
  if in.Mode != ModeAppend && in.NewStr == "" {
    return "", fmt.Errorf("new_str required for mode=%s", in.Mode)
  }
  if in.ExpectCount == 0 {
    in.ExpectCount = 1
  }
  p, err := safePath(in.Path)
  if err != nil {
    return "", err
  }

  // Read current (or empty)
  var cur []byte
  if b, err := os.ReadFile(p); err == nil {
    cur = b
  } else if !os.IsNotExist(err) {
    return "", err
  }

  oldHash := fileHash(cur)
  if in.ExpectHash != "" && in.ExpectHash != oldHash {
    res, _ := json.Marshal(EditResult{
      Changed: false, OldHash: oldHash, NewHash: oldHash, Message: "expect_hash_mismatch",
    })
    return string(res), nil
  }

  newContent := cur
  replacements := 0

  switch in.Mode {
  case ModeWriteFull:
    newContent = []byte(in.NewStr)

  case ModeAppend:
    if in.EnsureNL && len(cur) > 0 && cur[len(cur)-1] != '\n' {
      newContent = append(cur, '\n')
    }
    newContent = append(newContent, []byte(in.NewStr)...)

  case ModeReplaceOnce:
    curStr := string(cur)
    count := strings.Count(curStr, in.OldStr)
    if count != in.ExpectCount {
      res, _ := json.Marshal(EditResult{
        Changed: false, OldHash: oldHash, NewHash: oldHash,
        Message: fmt.Sprintf("match_count_mismatch: have=%d expect=%d", count, in.ExpectCount),
      })
      return string(res), nil
    }
    newStr := strings.Replace(curStr, in.OldStr, in.NewStr, in.ExpectCount)
    if newStr == curStr {
      res, _ := json.Marshal(EditResult{
        Changed: false, OldHash: oldHash, NewHash: oldHash, Message: "no_change",
      })
      return string(res), nil
    }
    newContent = []byte(newStr)
    replacements = in.ExpectCount
  }

  // Only write if changed
  newHash := fileHash(newContent)
  if newHash != oldHash {
    if dir := filepath.Dir(p); dir != "." {
      if err := os.MkdirAll(dir, 0755); err != nil {
        return "", err
      }
    }
    if err := os.WriteFile(p, newContent, 0644); err != nil {
      return "", err
    }
  }

  res, _ := json.Marshal(EditResult{
    Changed: (newHash != oldHash),
    OldHash: oldHash,
    NewHash: newHash,
    Replacements: replacements,
    Message: func() string {
      if newHash == oldHash { return "no_change" }
      switch in.Mode {
      case ModeWriteFull: return "wrote_full"
      case ModeAppend:    return "appended"
      default:            return "replaced"
      }
    }(),
  })
  return string(res), nil
}