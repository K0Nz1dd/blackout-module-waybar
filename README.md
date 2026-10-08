# Waybar Blackout Schedule Module

A lightweight Go-based custom module for [Waybar](https://github.com/Alexays/Waybar) that scrapes power outage schedules from the regional power distribution website and displays the current blackout status along with a detailed daily schedule in the tooltip.

## Features

- **Real-time status**: Shows whether power is currently **ON** or **OFF**.
- **Dynamic Tooltip**: Displays the complete grouped outage schedule for the day with a "Last updated" timestamp.
- **CSS Styling**: Provides distinct CSS classes (`on` and `off`) for custom styling in Waybar.
- **Efficient Matching**: Maps time slots directly using mathematically calculated index intervals.

---

## Installation & Building
### 1. Prerequisites
Make sure you have [Go](https://go.dev/) installed on your system (version 1.18+ recommended).
### 2. Clone the Repository
```bash
git clone [https://github.com/your-username/blackout-module-waybar.git](https://github.com/your-username/blackout-module-waybar.git)
cd blackout-module-waybar
```
### 3. Build the Binary
#### Compile the application executable from the project root:
```bash
go build -o blackout ./cmd/blackout
```
### 4. Make it Executable & Move to PATH
#### Grant execution permissions to the generated binary and move it to a local binary directory (e.g., ~/.local/bin):
```bash
chmod +x blackout
mv blackout ~/.local/bin/
```
### Waybar Integration
#### 1. Module Configuration (~/.config/waybar/config)
#### Add the custom module definition to your Waybar config file:
```bash
  "custom/blackout": {
      "exec": "~/.config/scripts/blackout",
      "return-type": "json",
    "interval": 3600,
    "format": "{icon}",
    "format-icons": {
        "on": "󰌵 ",
        "off": "󰹏 "
      },
    "max-length": 350,
    "tooltip": true,
    "on-click": "~/.config/scripts/blackout"
    },
}
```
#### Include "custom/blackout" in your main layout structure (e.g., modules-right):
```bash
"modules-right": [
    "custom/blackout",
    "clock",
    "tray"
]
```
### 2. Styling (~/.config/waybar/style.css)
#### Add CSS rules to ~/.config/waybar/style.css to style the module based on power availability:
```bash
#custom-blackout {
    padding: 0 10px;
    font-weight: bold;
}
/* Power is ON */
#custom-blackout.on {
    color: #a6e3a1;
    background-color: #1e1e2e;
}
/* Power is OFF */
#custom-blackout.off {
    color: #f38ba8;
    background-color: #313244;
    animation: blink 2s infinite;
}
```