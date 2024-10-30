package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon
#include <Carbon/Carbon.h>
#include <CoreFoundation/CoreFoundation.h>

// 添加事件类型定义
#define MY_CGEventFlagsChanged 12

extern void goCallback(CGEventRef event, CGEventType type, CGKeyCode keyCode, CGEventFlags flags);

static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    CGKeyCode keyCode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
    CGEventFlags flags = CGEventGetFlags(event);

    // 检查是否是 Shift 或 ESC 键
    if (type == kCGEventKeyDown || type == kCGEventKeyUp || type == kCGEventFlagsChanged) {
        goCallback(event, type, keyCode, flags);
    }
    return event;
}

static void RunEventTap() {  // 添加 static 关键字
    CGEventMask eventMask = (CGEventMaskBit(kCGEventKeyDown) |
                            CGEventMaskBit(kCGEventKeyUp) |
                            CGEventMaskBit(kCGEventFlagsChanged));

    CFMachPortRef tap = CGEventTapCreate(
        kCGHIDEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        eventMask,
        eventCallback,
        NULL
    );

    if (!tap) {
        printf("Failed to create event tap. Please check accessibility permissions.\n");
        return;
    }

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(
        kCFAllocatorDefault,
        tap,
        0
    );

    CFRunLoopAddSource(
        CFRunLoopGetCurrent(),
        runLoopSource,
        kCFRunLoopCommonModes
    );

    CGEventTapEnable(tap, true);
    CFRunLoopRun();
}
*/
import "C"

var (
	shiftPressed       bool
	lastShiftPressTime int64
)

type Config struct {
	PrimaryIM   string `json:"primary_im"`
	SecondaryIM string `json:"secondary_im"`
}

const (
	version            = "1.0.0"
	defaultPrimaryIM   = "com.apple.keylayout.ABC"
	defaultSecondaryIM = "com.tencent.inputmethod.wetype.pinyin"
)

var config Config

func loadConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "mac-vim-switch")
	configFile := filepath.Join(configDir, "config.json")

	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}

		config = Config{
			PrimaryIM:   defaultPrimaryIM,
			SecondaryIM: defaultSecondaryIM,
		}

		return saveConfig(configFile)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &config)
}

func backupConfig(configFile string) error {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	backupFile := configFile + ".backup"
	return os.WriteFile(backupFile, data, 0644)
}

func saveConfig(configFile string) error {
	// 先备份当前配置
	if _, err := os.Stat(configFile); err == nil {
		if err := backupConfig(configFile); err != nil {
			log.Printf("Warning: Failed to backup config: %v\n", err)
		}
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func listAvailableInputMethods() ([]string, error) {
	cmd := exec.Command("macism")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get input methods: %v (is macism installed?)", err)
	}

	methods := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(methods) == 0 {
		return nil, fmt.Errorf("no input methods found")
	}
	return methods, nil
}

func getCurrentInputMethod() (string, error) {
	cmd := exec.Command("macism")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current input method: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func switchToInputMethod(im string) error {
	cmd := exec.Command("macism", im)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to switch input method: %v", err)
	}
	return nil
}

func checkInputMethodExists(im string) bool {
	cmd := exec.Command("macism", im)
	return cmd.Run() == nil
}

func setupLogging() *os.File {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	logFile, err := os.OpenFile(homeDir+"/.mac-vim-switch.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	return logFile
}

func validateInputMethod(im string) error {
	methods, err := listAvailableInputMethods()
	if err != nil {
		return err
	}

	for _, method := range methods {
		if method == im {
			return nil
		}
	}
	return fmt.Errorf("input method '%s' not found. Use 'mac-vim-switch list' to see available methods", im)
}

func healthCheck() error {
	// 检查 macism 否可用
	if _, err := exec.LookPath("macism"); err != nil {
		return fmt.Errorf("macism not found: %v", err)
	}

	// 检查配置文件权限
	configFile := filepath.Join(os.Getenv("HOME"), ".config", "mac-vim-switch", "config.json")
	if _, err := os.Stat(configFile); err == nil {
		// 检查是否可读写
		if err := unix.Access(configFile, unix.R_OK|unix.W_OK); err != nil {
			return fmt.Errorf("config file permission error: %v", err)
		}
	}

	// 检查日志文件权限
	logFile := filepath.Join(os.Getenv("HOME"), ".mac-vim-switch.log")
	if _, err := os.Stat(logFile); err == nil {
		if err := unix.Access(logFile, unix.R_OK|unix.W_OK); err != nil {
			return fmt.Errorf("log file permission error: %v", err)
		}
	}

	return nil
}

//export goCallback
func goCallback(event C.CGEventRef, eventType C.CGEventType, keyCode C.CGKeyCode, flags C.CGEventFlags) {
	now := time.Now().UnixNano()

	// 记录所有按键事件
	eventTypeStr := "Unknown"
	switch eventType {
	case C.kCGEventKeyDown:
		eventTypeStr = "KeyDown"
	case C.kCGEventKeyUp:
		eventTypeStr = "KeyUp"
	case 12: // kCGEventFlagsChanged
		eventTypeStr = "FlagsChanged"
	}

	log.Printf("Key Event: type=%s, keyCode=0x%x, flags=0x%x\n",
		eventTypeStr, keyCode, flags)

	// 如果是普通按键事件（非修饰键），重置 shift 状态
	if eventType == C.kCGEventKeyDown {
		shiftPressed = false
		return
	}

	switch keyCode {
	case 0x35: // ESC
		if eventType == C.kCGEventKeyDown {
			log.Println("ESC key pressed, switching to primary input method")
			if err := switchToInputMethod(config.PrimaryIM); err != nil {
				log.Printf("Error switching to primary input method: %v\n", err)
			}
		}
	case 0x38, 0x3C: // Left or Right Shift
		if eventType == 12 { // FlagsChanged event
			isShiftDown := flags&C.kCGEventFlagMaskShift != 0
			hasOtherModifiers := (flags & (C.kCGEventFlagMaskCommand |
				C.kCGEventFlagMaskControl |
				C.kCGEventFlagMaskAlternate |
				C.kCGEventFlagMaskSecondaryFn)) != 0

			log.Printf("Shift state changed - isDown: %v, hasOtherModifiers: %v, flags: 0x%x\n",
				isShiftDown, hasOtherModifiers, flags)

			if hasOtherModifiers {
				shiftPressed = false
				return
			}

			if isShiftDown {
				if !shiftPressed {
					shiftPressed = true
					lastShiftPressTime = now
				}
			} else if shiftPressed {
				shiftPressed = false
				if now-lastShiftPressTime < 300*1000000 { // 300ms
					current, err := getCurrentInputMethod()
					if err != nil {
						log.Printf("Error getting current input method: %v\n", err)
						return
					}

					targetIM := config.SecondaryIM
					if current == config.SecondaryIM {
						targetIM = config.PrimaryIM
					}

					if err := switchToInputMethod(targetIM); err != nil {
						log.Printf("Error switching input method: %v\n", err)
					}
				}
			}
		}
	}
}

func main() {
	if err := loadConfig(); err != nil {
		log.Printf("Error loading config: %v\n", err)
		log.Println("Using default settings")
		config = Config{
			PrimaryIM:   defaultPrimaryIM,
			SecondaryIM: defaultSecondaryIM,
		}
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			methods, err := listAvailableInputMethods()
			if err != nil {
				log.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Available input methods:")
			for _, method := range methods {
				fmt.Println(method)
			}
			return
		case "config":
			if len(os.Args) == 2 {
				fmt.Printf("Current configuration:\n")
				fmt.Printf("Primary input method: %s\n", config.PrimaryIM)
				fmt.Printf("Secondary input method: %s\n", config.SecondaryIM)
				return
			}
			if len(os.Args) != 4 {
				fmt.Println("Usage: mac-vim-switch config [primary|secondary] <input-method-id>")
				fmt.Println("Example: mac-vim-switch config secondary com.apple.inputmethod.SCIM.ITABC")
				fmt.Println("\nUse 'mac-vim-switch list' to see available input methods")
				os.Exit(1)
			}
			switch os.Args[2] {
			case "primary":
				if err := validateInputMethod(os.Args[3]); err != nil {
					log.Printf("Error: %v\n", err)
					os.Exit(1)
				}
				config.PrimaryIM = os.Args[3]
			case "secondary":
				if err := validateInputMethod(os.Args[3]); err != nil {
					log.Printf("Error: %v\n", err)
					os.Exit(1)
				}
				config.SecondaryIM = os.Args[3]
			default:
				log.Printf("Unknown config option: %s\n", os.Args[2])
				os.Exit(1)
			}
			if err := saveConfig(filepath.Join(os.Getenv("HOME"), ".config", "mac-vim-switch", "config.json")); err != nil {
				log.Printf("Error saving config: %v\n", err)
				os.Exit(1)
			}
			return
		case "version", "--version", "-v":
			fmt.Printf("mac-vim-switch version %s\n", version)
			return
		case "health", "doctor":
			if err := healthCheck(); err != nil {
				fmt.Printf("Health check failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("All systems operational!")
			return
		}
		current, err := getCurrentInputMethod()
		if err != nil {
			log.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		switch os.Args[1] {
		case "esc":
			if err := switchToInputMethod(config.PrimaryIM); err != nil {
				log.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		case "shift":
			if current == config.PrimaryIM {
				if err := switchToInputMethod(config.SecondaryIM); err != nil {
					log.Printf("Error: %v\n", err)
					os.Exit(1)
				}
			} else {
				if err := switchToInputMethod(config.PrimaryIM); err != nil {
					log.Printf("Error: %v\n", err)
					os.Exit(1)
				}
			}
		}
		return
	}

	// 守护进程模式
	logFile := setupLogging()
	defer logFile.Close()

	log.Println("Starting mac-vim-switch daemon...")

	if _, err := exec.LookPath("macism"); err != nil {
		log.Fatal("Error: macism is not installed. Please install it first:\nbrew tap laishulu/macism\nbrew install macism")
	}

	if !checkInputMethodExists(config.SecondaryIM) {
		log.Println("Warning: Secondary input method not found.")
		log.Println("Please use 'macism' command to check available input methods.")
		log.Println("Current setting will only work with", config.PrimaryIM)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动事件监听
	go C.RunEventTap()

	log.Println("Daemon is running...")
	<-sigChan
	log.Println("Shutting down...")
}
