package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

type ChangeItemType string

const (
	ChangeContext ChangeItemType = "Context"
	ChangeCluster ChangeItemType = "Cluster"
	ChangeUser    ChangeItemType = "User"
)

type ChangeAction string

const (
	ActionAdd    ChangeAction = "NEW"
	ActionModify ChangeAction = "CHANGED"
	ActionDelete ChangeAction = "DELETED"
)

// ChangeItem is a flattened representation of a single change for the interactive selection menu.
type ChangeItem struct {
	Type   ChangeItemType
	Action ChangeAction
	Name   string
	Value  interface{}
}

// ToChangeItems flattens a KubeconfigDiff into a slice of ChangeItems.
func (d KubeconfigDiff) ToChangeItems() []ChangeItem {
	var items []ChangeItem

	// Contexts
	for _, x := range d.ContextsAdded {
		items = append(items, ChangeItem{Type: ChangeContext, Action: ActionAdd, Name: x.Name, Value: x})
	}

	for _, x := range d.ContextsModified {
		items = append(items, ChangeItem{Type: ChangeContext, Action: ActionModify, Name: x.Name, Value: x})
	}

	for _, name := range d.ContextsDeleted {
		items = append(items, ChangeItem{Type: ChangeContext, Action: ActionDelete, Name: name, Value: name})
	}

	// Clusters
	for _, x := range d.ClustersAdded {
		items = append(items, ChangeItem{Type: ChangeCluster, Action: ActionAdd, Name: x.Name, Value: x})
	}

	for _, x := range d.ClustersModified {
		items = append(items, ChangeItem{Type: ChangeCluster, Action: ActionModify, Name: x.Name, Value: x})
	}

	for _, name := range d.ClustersDeleted {
		items = append(items, ChangeItem{Type: ChangeCluster, Action: ActionDelete, Name: name, Value: name})
	}

	// Users
	for _, x := range d.UsersAdded {
		items = append(items, ChangeItem{Type: ChangeUser, Action: ActionAdd, Name: x.Name, Value: x})
	}

	for _, x := range d.UsersModified {
		items = append(items, ChangeItem{Type: ChangeUser, Action: ActionModify, Name: x.Name, Value: x})
	}

	for _, name := range d.UsersDeleted {
		items = append(items, ChangeItem{Type: ChangeUser, Action: ActionDelete, Name: name, Value: name})
	}

	return items
}

func applyItem(filtered *KubeconfigDiff, item ChangeItem) {
	switch item.Type {
	case ChangeContext:
		switch item.Action {
		case ActionAdd:
			filtered.ContextsAdded = append(filtered.ContextsAdded, item.Value.(apiv1.NamedContext))
		case ActionModify:
			filtered.ContextsModified = append(filtered.ContextsModified, item.Value.(apiv1.NamedContext))
		case ActionDelete:
			filtered.ContextsDeleted = append(filtered.ContextsDeleted, item.Value.(string))
		}
	case ChangeCluster:
		switch item.Action {
		case ActionAdd:
			filtered.ClustersAdded = append(filtered.ClustersAdded, item.Value.(apiv1.NamedCluster))
		case ActionModify:
			filtered.ClustersModified = append(filtered.ClustersModified, item.Value.(apiv1.NamedCluster))
		case ActionDelete:
			filtered.ClustersDeleted = append(filtered.ClustersDeleted, item.Value.(string))
		}
	case ChangeUser:
		switch item.Action {
		case ActionAdd:
			filtered.UsersAdded = append(filtered.UsersAdded, item.Value.(apiv1.NamedAuthInfo))
		case ActionModify:
			filtered.UsersModified = append(filtered.UsersModified, item.Value.(apiv1.NamedAuthInfo))
		case ActionDelete:
			filtered.UsersDeleted = append(filtered.UsersDeleted, item.Value.(string))
		}
	}
}

func runTUI(items []ChangeItem) (checked []bool, cancelled bool, err error) {
	checked = make([]bool, len(items))

	for i, item := range items {
		if item.Action == ActionAdd || item.Action == ActionModify {
			checked[i] = true
		} else {
			checked[i] = false
		}
	}

	cursor := 0

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, false, err
	}

	defer func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}()

	printMenu := func(firstTime bool) {
		if !firstTime {
			clearSequence := strings.Repeat("\033[A\r\033[K", len(items))
			fmt.Print(clearSequence)
		}

		for i, item := range items {
			cursorStr := "  "
			if i == cursor {
				cursorStr = " >"
			}

			checkStr := "[ ]"
			if checked[i] {
				checkStr = "[✓]"
			}

			var rawLabel string

			switch item.Action {
			case ActionAdd:
				rawLabel = "NEW"
			case ActionModify:
				rawLabel = "CHANGED"
			case ActionDelete:
				rawLabel = "DELETED"
			}

			paddedLabel := fmt.Sprintf("%-7s", rawLabel)

			var coloredLabel string

			switch item.Action {
			case ActionAdd:
				coloredLabel = "\033[32m" + paddedLabel + "\033[0m"
			case ActionModify:
				coloredLabel = "\033[33m" + paddedLabel + "\033[0m"
			case ActionDelete:
				coloredLabel = "\033[31m" + paddedLabel + "\033[0m"
			}

			if i == cursor {
				fmt.Printf("%s \033[1m%s %s %s: %s\033[0m\r\n", cursorStr, checkStr, coloredLabel, item.Type, item.Name)
			} else {
				fmt.Printf("%s %s %s %s: %s\r\n", cursorStr, checkStr, coloredLabel, item.Type, item.Name)
			}
		}
	}

	printMenu(true)

	buf := make([]byte, 3)

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, false, err
		}

		if n == 1 {
			b := buf[0]
			switch b {
			case 3: // Ctrl-C
				return nil, true, nil
			case 27: // Esc
				return nil, true, nil
			case 13, 10: // Enter
				return checked, false, nil
			case ' ': // Spacebar
				checked[cursor] = !checked[cursor]

				printMenu(false)
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 65: // Up Arrow
				if cursor > 0 {
					cursor--

					printMenu(false)
				}
			case 66: // Down Arrow
				if cursor < len(items)-1 {
					cursor++

					printMenu(false)
				}
			}
		}
	}
}

// selectChanges prompts the user interactively to select which changes to apply.
func selectChanges(diff KubeconfigDiff, originalPath, tempPath string) (KubeconfigDiff, error) {
	items := diff.ToChangeItems()
	if len(items) == 0 {
		return KubeconfigDiff{}, nil
	}

	fmt.Println("\n\033[1;36m❖ ksw (kubeconfig-switcher): Exiting session.\033[0m")
	fmt.Println("Detected modifications in your temporary session kubeconfig.")
	fmt.Printf("  Original:  %s\n", originalPath)
	fmt.Printf("  Temporary: %s\n\n", tempPath)
	fmt.Println("Select which changes you want to apply back to the original kubeconfig:")
	fmt.Println("(\033[32mNEW\033[0m and \033[33mCHANGED\033[0m items are pre-selected. Use Up/Down arrows to move, Space to toggle, Enter to confirm, Esc to cancel.)")
	fmt.Println()

	checked, cancelled, err := runTUI(items)
	if err != nil {
		return KubeconfigDiff{}, err
	}

	if cancelled {
		fmt.Println("Merge discarded. No changes applied.")

		return KubeconfigDiff{}, nil
	}

	var filtered KubeconfigDiff

	hasAnyChecked := false

	for i, isChecked := range checked {
		if isChecked {
			hasAnyChecked = true

			applyItem(&filtered, items[i])
		}
	}

	if !hasAnyChecked {
		fmt.Println("No changes selected to merge.")

		return KubeconfigDiff{}, nil
	}

	return filtered, nil
}
