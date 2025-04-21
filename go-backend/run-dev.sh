#!/bin/bash
echo "$SUDO_PASSWORD" | sudo -E -S $(which go) run .