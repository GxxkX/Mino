import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../core/theme/app_theme.dart';
import '../../core/providers/auth_provider.dart';
import '../../features/recording/models/audio_config.dart';
import '../../features/recording/providers/recording_provider.dart';
import '../../features/settings/providers/settings_provider.dart';
import '../widgets/recording_button.dart';
import '../widgets/transcript_display.dart';
import 'settings_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _checkRecordingPermission();
    });
  }

  Future<void> _checkRecordingPermission() async {
    final recordingProvider = context.read<RecordingProvider>();
    final settings = context.read<SettingsProvider>();
    final hasPermission = await recordingProvider.checkPermission();
    if (!hasPermission && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            settings.t('Microphone permission required', '需要麦克风权限'),
            style: AppTheme.bodyStyle(fontSize: 13),
          ),
          backgroundColor: AppTheme.warning,
        ),
      );
    }
  }

  Future<void> _toggleRecording() async {
    final authProvider = context.read<AuthProvider>();
    final recordingProvider = context.read<RecordingProvider>();
    final settings = context.read<SettingsProvider>();

    if (recordingProvider.isRecording) {
      await recordingProvider.stopRecording();
    } else {
      // startRecording clears the transcript internally, so no need to reset first
      await recordingProvider.startRecording(
        authProvider,
        gain: settings.recordingGain,
        language: settings.language,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.background,
      appBar: AppBar(
        backgroundColor: AppTheme.background,
        elevation: 0,
        title: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            ClipRRect(
              borderRadius: BorderRadius.circular(6),
              child: Image.asset('assets/logo.png', width: 28, height: 28),
            ),
            const SizedBox(width: 8),
            Text('Mino', style: AppTheme.headingStyle(fontSize: 24)),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings, color: AppTheme.textSecondary),
            onPressed: () {
              Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const SettingsScreen()),
              );
            },
          ),
        ],
      ),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(AppTheme.spaceMd),
          child: Column(
            children: [
              Consumer<AuthProvider>(
                builder: (context, auth, _) {
                  return Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: AppTheme.spaceSm,
                      vertical: AppTheme.spaceXs,
                    ),
                    child: Row(
                      children: [
                        Container(
                          width: 32,
                          height: 32,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            color: AppTheme.cta.withOpacity(0.15),
                          ),
                          child: const Icon(Icons.person,
                              size: 18, color: AppTheme.cta),
                        ),
                        const SizedBox(width: AppTheme.spaceSm),
                        Text(
                          auth.username ?? 'User',
                          style: AppTheme.bodyStyle(
                              fontWeight: FontWeight.w500),
                        ),
                        const Spacer(),
                        Consumer<RecordingProvider>(
                          builder: (context, recording, _) {
                            if (!recording.isRecording) {
                              return const SizedBox.shrink();
                            }
                            return Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: AppTheme.spaceSm,
                                vertical: AppTheme.spaceXs,
                              ),
                              decoration: BoxDecoration(
                                color: AppTheme.error.withOpacity(0.1),
                                borderRadius: BorderRadius.circular(
                                    AppTheme.radiusSm),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Container(
                                    width: 6,
                                    height: 6,
                                    decoration: const BoxDecoration(
                                      shape: BoxShape.circle,
                                      color: AppTheme.error,
                                    ),
                                  ),
                                  const SizedBox(width: AppTheme.spaceXs),
                                  Text(
                                    'REC',
                                    style: AppTheme.bodyStyle(
                                      fontSize: 10,
                                      fontWeight: FontWeight.w600,
                                      color: AppTheme.error,
                                    ),
                                  ),
                                ],
                              ),
                            );
                          },
                        ),
                      ],
                    ),
                  );
                },
              ),
              const SizedBox(height: AppTheme.spaceMd),
              // Success banner
              Consumer2<RecordingProvider, SettingsProvider>(
                builder: (context, recording, settings, _) {
                  if (recording.state != RecordingState.completed) {
                    return const SizedBox.shrink();
                  }
                  return Container(
                    width: double.infinity,
                    margin:
                        const EdgeInsets.only(bottom: AppTheme.spaceSm),
                    padding: const EdgeInsets.symmetric(
                      horizontal: AppTheme.spaceMd,
                      vertical: AppTheme.spaceSm,
                    ),
                    decoration: BoxDecoration(
                      color: AppTheme.success.withOpacity(0.12),
                      borderRadius:
                          BorderRadius.circular(AppTheme.radiusSm),
                      border: Border.all(
                          color: AppTheme.success.withOpacity(0.3)),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.check_circle_outline,
                            color: AppTheme.success, size: 16),
                        const SizedBox(width: AppTheme.spaceSm),
                        Text(
                          settings.t('Recording uploaded successfully',
                              '录音已上传成功'),
                          style: AppTheme.bodyStyle(
                            fontSize: 13,
                            color: AppTheme.success,
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
              Expanded(
                child: Consumer<RecordingProvider>(
                  builder: (context, recording, _) {
                    return TranscriptDisplay(
                      transcript: recording.transcript,
                      isRecording: recording.isRecording,
                      isProcessing:
                          recording.state == RecordingState.processing,
                    );
                  },
                ),
              ),
              const SizedBox(height: AppTheme.spaceLg),
              Consumer<RecordingProvider>(
                builder: (context, recording, _) {
                  final isLoading =
                      recording.state == RecordingState.connecting ||
                          recording.state == RecordingState.processing;
                  return RecordingButton(
                    isRecording: recording.isRecording,
                    isLoading: isLoading,
                    onPressed: _toggleRecording,
                  );
                },
              ),
              const SizedBox(height: AppTheme.spaceSm),
              Consumer<RecordingProvider>(
                builder: (context, recording, _) {
                  if (recording.error.isNotEmpty) {
                    return Padding(
                      padding:
                          const EdgeInsets.only(bottom: AppTheme.spaceXs),
                      child: Text(
                        recording.error,
                        style: AppTheme.bodyStyle(
                            fontSize: 12, color: AppTheme.error),
                      ),
                    );
                  }
                  return const SizedBox.shrink();
                },
              ),
              Consumer2<RecordingProvider, SettingsProvider>(
                builder: (context, recording, settings, _) {
                  if (recording.isRecording) {
                    return TextButton(
                      onPressed: () => recording.cancelRecording(),
                      child: Text(
                        settings.t('Cancel', '取消'),
                        style: AppTheme.bodyStyle(
                            color: AppTheme.textSecondary),
                      ),
                    );
                  }
                  return const SizedBox.shrink();
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
}
