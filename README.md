# sprt - Spotify CLI Client

sprt is a command-line interface (CLI) tool for interacting with Spotify. It allows you to authenticate with Spotify, get information about your currently playing track, and display synchronized lyrics for the current track.

## Features

- Authenticate with Spotify using the Authorization Code Flow
- Store authentication tokens securely in a local file
- Get information about your currently playing track
- Display synchronized lyrics for the currently playing track
- Automatic token refresh when expired

## Installation

### Prerequisites

- Go 1.16 or higher (for building from source)
- A Spotify Developer account
- A registered Spotify application with a client ID and client secret

### Using instl.sh (Recommended)

You can install sprt with a single command:

```bash
curl -sSL https://instl.sh/muhadif/sprt | bash
```

This will automatically download and install the latest version of sprt for your platform.

### Using Make (Linux and macOS)

1. Clone the repository:

```bash
git clone https://github.com/muhadif/sprt.git
cd sprt
```

2. Install the application:

```bash
make install
```

This will build the application and install it to `/usr/local/bin/sprt`.

### Using the Installation Script

1. Download the latest release for your platform (Linux or macOS)
2. Extract the archive:

```bash
tar -xzf sprt-[platform].tar.gz
```

3. Run the installation script:

```bash
sudo ./instl.sh
```

### Building from Source

1. Clone the repository:

```bash
git clone https://github.com/muhadif/sprt.git
cd sprt
```

2. Build the application:

```bash
go build -o sprt ./cmd/sprt
```

3. Move the binary to your PATH (optional):

```bash
sudo mv sprt /usr/local/bin/
```

## Usage

### Authentication

To initialize the authentication process:

```bash
sprt auth init
```

This will prompt you for your Spotify client ID and client secret, then display an authorization URL. Open this URL in your browser to authorize the application. After authorization, you will be redirected to a local callback URL, and the application will exchange the authorization code for an access token.

### Getting Currently Playing Track

To get information about your currently playing track:

```bash
sprt current
```

This will display the title, artist, and album of the currently playing track.

### Displaying Synchronized Lyrics

To display synchronized lyrics for the currently playing track:

```bash
sprt lyric pipe-lyric
```

This will fetch lyrics from lrclib.net and display them synchronized with the music. Press Ctrl+C to stop the lyrics display.

## Developer Guide

### Setting Up Spotify Integration

To integrate your application with Spotify:

1. Create a Spotify Developer account at [developer.spotify.com](https://developer.spotify.com/)
2. Create a new application in the [Spotify Developer Dashboard](https://developer.spotify.com/dashboard/applications)
3. Set the Redirect URI to `http://127.0.0.1:8080/callback`
4. Note your Client ID and Client Secret
5. Use these credentials when running `sprt auth init`

### API Scopes

sprt uses the following Spotify API scopes:
- `user-read-currently-playing`: Required to get information about the currently playing track

### Adding New Features

To add new features to sprt:

1. Define new use cases in the `domain/usecase` package
2. Implement any required repositories in the `infrastructure/persistence` package
3. Add new commands in the `cmd/sprt/cmd` package
4. Update the README.md with documentation for the new features

### Release Process

sprt uses [GoReleaser](https://goreleaser.com/) and GitHub Actions to automate the release process:

1. Update the code and commit your changes
2. Create and push a new tag with the version number:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. The GitHub Actions workflow will automatically:
   - Build binaries for Linux and macOS (both amd64 and arm64)
   - Create archives with the binary, README, and installation script
   - Generate checksums for verification
   - Create a GitHub release with the generated assets
   - Generate release notes from commit messages

To test the release process locally (without publishing):

```bash
# Install GoReleaser if you don't have it
go install github.com/goreleaser/goreleaser@latest

# Run GoReleaser in snapshot mode (no actual release)
goreleaser release --snapshot --clean
```

## Architecture

The application is built using clean architecture principles, with the following layers:

- **Domain**: Contains the core business logic and entities
  - **Entity**: Defines the data structures
  - **Repository**: Defines the interfaces for data access
  - **Usecase**: Implements the business rules

- **Infrastructure**: Contains the implementation details
  - **Persistence**: Implements the repository interfaces
  - **Auth**: Handles authentication with external services

- **Interfaces**: Contains the user interfaces
  - **CLI**: Implements the command-line interface
  - **HTTP**: Implements the HTTP server for callbacks

## Troubleshooting

### Authentication Issues

If you encounter authentication issues:
- Ensure your Client ID and Client Secret are correct
- Check that your Redirect URI is set correctly in the Spotify Developer Dashboard
- Try running `sprt auth init` again to re-authenticate

### No Track Playing

If you get a "No track currently playing" message:
- Make sure you have a track playing on Spotify
- Check that your Spotify account is active and not in offline mode

### Lyrics Not Found

If lyrics are not found for a track:
- The track may not have lyrics available in the lrclib.net database
- Check if the artist and track names are correct

## Linux Desktop Integration

### GNOME Shell Integration with Executor

sprt writes the current lyric to `/tmp/current-lyric.txt` when you run the `sprt lyric pipe-lyric` command. You can display these lyrics on your GNOME desktop using the [Executor extension](https://extensions.gnome.org/extension/2932/executor/).

To set up:

1. Install the Executor extension from [GNOME Extensions](https://extensions.gnome.org/extension/2932/executor/)
2. Open the Executor settings
3. Add a new command with the following settings:
   - Command: `cat /tmp/current-lyric.txt`
   - Refresh interval: 1 second
   - Display options: Configure as desired (position, font, etc.)

This will display the currently playing lyric on your desktop, synchronized with your music.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Spotify Web API](https://developer.spotify.com/documentation/web-api/)
- [OAuth 2.0 Authorization Code Flow](https://developer.spotify.com/documentation/web-api/tutorials/code-flow)
- [lrclib.net](https://lrclib.net/) for providing synchronized lyrics
