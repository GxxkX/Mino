import 'package:flutter/material.dart';
import '../../core/theme/app_theme.dart';

class TranscriptDisplay extends StatelessWidget {
  final String transcript;
  final bool isRecording;
  final bool isProcessing;

  const TranscriptDisplay({
    super.key,
    required this.transcript,
    this.isRecording = false,
    this.isProcessing = false,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(AppTheme.spaceMd),
      decoration: BoxDecoration(
        color: AppTheme.primary,
        borderRadius: BorderRadius.circular(AppTheme.radiusMd),
        boxShadow: AppTheme.shadowSm,
        border: Border.all(
          color: isRecording
              ? AppTheme.cta.withOpacity(0.2)
              : AppTheme.secondary,
          width: 0.5,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header row
          Row(
            children: [
              if (isRecording) ...[
                _PulsingDot(),
                const SizedBox(width: AppTheme.spaceSm),
              ],
              Text(
                isRecording ? 'Recording...' : isProcessing ? 'Processing...' : 'Transcript',
                style: AppTheme.bodyStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w600,
                  color: AppTheme.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: AppTheme.spaceSm),
          // Transcript content
          Expanded(
            child: SingleChildScrollView(
              child: Text(
                transcript.isEmpty
                    ? 'Tap the button to start recording'
                    : transcript,
                style: AppTheme.bodyStyle(
                  fontSize: 14,
                  color: transcript.isEmpty
                      ? AppTheme.textSecondary
                      : AppTheme.text,
                  fontWeight:
                      transcript.isEmpty ? FontWeight.w400 : FontWeight.w400,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Animated pulsing dot indicator for recording state
class _PulsingDot extends StatefulWidget {
  @override
  State<_PulsingDot> createState() => _PulsingDotState();
}

class _PulsingDotState extends State<_PulsingDot>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 800),
    );
    _animation = Tween<double>(begin: 0.4, end: 1.0).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
    _controller.repeat(reverse: true);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FadeTransition(
      opacity: _animation,
      child: Container(
        width: 8,
        height: 8,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: AppTheme.error,
          boxShadow: [
            BoxShadow(
              color: AppTheme.error.withOpacity(0.4),
              blurRadius: 4,
              spreadRadius: 1,
            ),
          ],
        ),
      ),
    );
  }
}
