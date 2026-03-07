# Watch client

This document describes the Mino smartwatch application, built with
Flutter for Wear OS devices. The watch app focuses on quick voice
capture with real-time transcription display.

## Technology stack

The watch application uses the following technologies.

| Technology | Purpose |
|------------|---------|
| Flutter (Dart 3.0+) | Cross-platform UI framework |
| Provider | State management (ChangeNotifier pattern) |
| web_socket_channel | Real-time audio streaming |
| record | Audio recording (16kHz mono) |
| sqflite | Local SQLite database for offline storage |
| shared_preferences | Token and settings persistence |

## Project structure

```
watch/
├── android/
│   └── app/
│       ├── build.gradle.kts       # Android config (SDK 25-34)
│       └── src/main/
│           ├── AndroidManifest.xml # Permissions
│           └── kotlin/.../
│               └── MainActivity.kt
├── assets/
│   └── icons/
├── lib/
│   ├── main.dart                  # App entry, MultiProvider setup
│   ├── app/
│   │   ├── screens/
│   │   │   ├── home_screen.dart   # Recording UI + transcript
│   │   │   ├── login_screen.dart  # Authentication
│   │   │   ├── permission_screen.dart
│   │   │   └── settings_screen.dart
│   │   └── widgets/
│   │       ├── recording_button.dart
│   │       └── transcript_display.dart
│   ├── core/
│   │   ├── constants/
│   │   │   └── app_config.dart    # API URLs, audio params
│   │   ├── providers/
│   │   │   └── auth_provider.dart
│   │   ├── services/
│   │   │   ├── auth_service.dart
│   │   │   └── database_service.dart
│   │   └── theme/
│   │       └── app_theme.dart
│   ├── features/
│   │   ├── recording/
│   │   │   ├── models/
│   │   │   │   └── audio_config.dart
│   │   │   ├── providers/
│   │   │   │   └── recording_provider.dart
│   │   │   └── services/
│   │   │       ├── audio_recording_service.dart
│   │   │       └── network_service.dart
│   │   └── settings/
│   │       └── providers/
│   │           └── settings_provider.dart
│   └── shared/
│       ├── providers/
│       │   └── connection_provider.dart
│       └── widgets/
│           └── battery_indicator.dart
└── pubspec.yaml
```

## App flow

The watch app follows a linear flow optimized for the small screen.

1. **Login**: Enter username and password to authenticate with the
   backend. Tokens are stored in `SharedPreferences`.
2. **Permission check**: Request microphone permission if not already
   granted.
3. **Home (recording)**: The main screen with a large recording button
   and real-time transcript display.
4. **Settings**: Account information, recording preferences, and sign
   out.

## Authentication

The watch app authenticates using the same REST API as other clients.
It calls `POST /v1/auth/signin` with username and password, then
stores the access token and refresh token in `SharedPreferences`.

The `AuthProvider` manages authentication state and exposes the current
user information to the widget tree through the Provider pattern.

## Recording

The recording feature is the core functionality of the watch app.

### Audio configuration

Audio is captured with the following settings, defined in
`lib/features/recording/models/audio_config.dart`.

| Parameter | Value |
|-----------|-------|
| Sample rate | 16,000 Hz |
| Channels | Mono |
| Encoding | PCM16 / AAC-LC |
| Bitrate | 32 kbps |
| Chunk interval | 100 ms |

### Recording flow

The `RecordingProvider` manages the complete recording lifecycle.

1. Establish a WebSocket connection to the backend at
   `/v1/ws/audio?token=<jwt_token>`.
2. Start audio capture using the `record` package.
3. Stream audio chunks every 100 ms through the WebSocket.
4. Receive and display real-time transcript updates from the backend.
5. Stop recording when the user taps the button or the maximum
   duration is reached.
6. Receive the final processing result (title, summary, action items,
   memories).

### Recording states

The app tracks recording state through the `RecordingState` enum.

| State | Description |
|-------|-------------|
| `idle` | Ready to record |
| `connecting` | Establishing WebSocket connection |
| `recording` | Actively recording and streaming |
| `processing` | Recording stopped, waiting for AI processing |
| `completed` | Processing finished |
| `error` | An error occurred |

## Offline support

When the network is unavailable, the watch app stores recordings
locally using SQLite.

The `DatabaseService` manages two SQLite tables.

| Table | Purpose |
|-------|---------|
| `conversations` | Cached conversation metadata |
| `offline_recordings` | Recordings pending upload |

When connectivity returns, the `SyncProvider` detects the network
change and uploads pending recordings to the backend. After successful
upload, local records are marked as synced.

## Settings

The `SettingsProvider` manages user preferences stored in
`SharedPreferences`.

| Setting | Description |
|---------|-------------|
| Audio quality | Recording quality level |
| Max duration | Maximum recording length |
| Offline mode | Enable/disable offline recording |
| Auto sync | Automatically sync when online |

## UI components

### Recording button

A circular animated button (`recording_button.dart`) that toggles
between microphone and stop icons. The button animates during recording
to provide visual feedback.

### Transcript display

A scrollable text area (`transcript_display.dart`) that shows the
real-time transcription as the user speaks. Text updates arrive through
the WebSocket connection.

### Battery indicator

A small widget (`battery_indicator.dart`) that displays the current
battery level with color coding (green, yellow, red).

## Theme

The watch app uses a dark OLED theme defined in
`lib/core/theme/app_theme.dart`. The color palette matches the project
design system with dark backgrounds for power efficiency on OLED
displays.

## Android configuration

The watch app targets Wear OS devices with the following Android
configuration.

| Setting | Value |
|---------|-------|
| Namespace | `com.mino.watch` |
| Min SDK | 25 (Android 7.1) |
| Target SDK | 34 (Android 14) |
| Screen size | 400 x 492 dp |

### Required permissions

The `AndroidManifest.xml` declares the following permissions.

- `INTERNET` for network communication
- `RECORD_AUDIO` for microphone access
- `BLUETOOTH` and `BLUETOOTH_ADMIN` for connectivity
- `READ_EXTERNAL_STORAGE` and `WRITE_EXTERNAL_STORAGE` for offline
  file storage

## Backend connection

The watch app connects to the backend using URLs configured in
`lib/core/constants/app_config.dart`.

| Endpoint | URL |
|----------|-----|
| REST API | `http://<host>:8000/v1` |
| WebSocket | `ws://<host>:8000/v1/ws/audio` |

Update these values to point to your backend server before building
the app.

## Development

### Prerequisites

Install Flutter 3.0 or later with Dart SDK 3.0+.

### Run on device

```bash
cd watch
flutter run
```

### Build release APK

```bash
flutter build apk --release
```

The release APK can be installed directly on Wear OS devices or
distributed through your preferred channel.

## Next steps

For backend API details, see the [API reference](api-reference.md).
For setting up the backend that the watch connects to, see the
[Getting started](getting-started.md) guide.
