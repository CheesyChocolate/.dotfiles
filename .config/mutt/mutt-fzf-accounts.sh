#!/bin/bash

ACCOUNTS_DIR="/home/suren/.config/mutt/accounts"
SWITCH_FILE="/home/suren/.config/mutt/.mutt_switch"

# Get account name from fzf, redirecting I/O to the terminal
ACCOUNT_BASENAME=$(find "$ACCOUNTS_DIR" -maxdepth 1 -name "*.muttrc" -printf "%f\n" | sed 's/\.muttrc$//' | sort | fzf --prompt="Switch to account: ")

# If an account was selected, link the corresponding .muttrc file to the switch file
if [ -n "$ACCOUNT_BASENAME" ]; then
	ln -sf "$ACCOUNTS_DIR/$ACCOUNT_BASENAME.muttrc" "$SWITCH_FILE"
fi
