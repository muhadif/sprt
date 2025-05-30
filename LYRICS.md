# Lyrics Configuration

This document provides detailed information about the lyrics feature in sprt, including configuration options and animation types.

## Configuration File

The lyrics configuration is stored in the UI configuration file at `~/.sprt/ui_config.json`. This file is automatically created with default values when you first run sprt.

## Configuration Options

The lyrics configuration includes the following options:

### Lyric Display

- `width`: The width of the lyric display area (default: 80)
- `height`: The height of the lyric display area (default: 20)

### Current Line Style

The style for the currently playing lyric line:

- `foregroundColor`: The text color in hex format (default: "#00FF00" - green)
- `backgroundColor`: The background color in hex format (default: "" - transparent)
- `bold`: Whether the text is bold (default: true)
- `italic`: Whether the text is italic (default: false)
- `underline`: Whether the text is underlined (default: false)

### Other Line Style

The style for other lyric lines:

- `foregroundColor`: The text color in hex format (default: "#FFFFFF" - white)
- `backgroundColor`: The background color in hex format (default: "" - transparent)
- `bold`: Whether the text is bold (default: false)
- `italic`: Whether the text is italic (default: false)
- `underline`: Whether the text is underlined (default: false)

### Animation

- `enabled`: Whether animations are enabled (default: true)
- `type`: The type of animation (default: "fade")
  - Available types: "fade", "slide", "none"
- `durationMs`: The duration of the animation in milliseconds (default: 300)
- `fadeSteps`: The number of steps for fade animation (default: 5)
- `slideDistance`: The distance to slide in characters for slide animation (default: 3)

## Example Configuration

Here's an example of a complete UI configuration file:

```json
{
  "lyric": {
    "currentLineStyle": {
      "foregroundColor": "#00FF00",
      "backgroundColor": "",
      "bold": true,
      "italic": false,
      "underline": false
    },
    "otherLineStyle": {
      "foregroundColor": "#FFFFFF",
      "backgroundColor": "",
      "bold": false,
      "italic": false,
      "underline": false
    },
    "width": 80,
    "height": 20,
    "animation": {
      "enabled": true,
      "type": "fade",
      "durationMs": 300,
      "fadeSteps": 5,
      "slideDistance": 3
    }
  }
}
```

## Customizing the Configuration

You can customize the lyrics display by editing the `~/.sprt/ui_config.json` file. After making changes, restart sprt for the changes to take effect.

### Animation Types

1. **Fade**: Smoothly fades between lyric lines
   - Controlled by `fadeSteps` parameter
   - Higher values create smoother but slower fades

2. **Slide**: Slides lyric lines in from the side
   - Controlled by `slideDistance` parameter
   - Higher values create longer slides

3. **None**: Disables animations for instant transitions

## Troubleshooting

If your custom configuration causes issues:

1. Delete the `~/.sprt/ui_config.json` file to reset to defaults
2. Run sprt again, and a new default configuration file will be created