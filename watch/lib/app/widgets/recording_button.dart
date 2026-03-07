import 'package:flutter/material.dart';
import '../../core/theme/app_theme.dart';

class RecordingButton extends StatefulWidget {
  final bool isRecording;
  final bool isLoading;
  final VoidCallback onPressed;

  const RecordingButton({
    super.key,
    required this.isRecording,
    required this.isLoading,
    required this.onPressed,
  });

  @override
  State<RecordingButton> createState() => _RecordingButtonState();
}

class _RecordingButtonState extends State<RecordingButton>
    with SingleTickerProviderStateMixin {
  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    );
    _pulseAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
  }

  @override
  void didUpdateWidget(RecordingButton oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.isRecording && !oldWidget.isRecording) {
      _pulseController.repeat(reverse: true);
    } else if (!widget.isRecording && oldWidget.isRecording) {
      _pulseController.stop();
      _pulseController.reset();
    }
  }

  @override
  void dispose() {
    _pulseController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final buttonColor = widget.isRecording ? AppTheme.error : AppTheme.cta;

    return GestureDetector(
      onTap: widget.isLoading ? null : widget.onPressed,
      child: AnimatedBuilder(
        animation: _pulseAnimation,
        builder: (context, child) {
          final pulseValue =
              widget.isRecording ? _pulseAnimation.value * 8 : 0.0;
          return AnimatedContainer(
            duration: AppTheme.transitionNormal,
            curve: Curves.easeInOut,
            width: widget.isRecording ? 80 : 96,
            height: widget.isRecording ? 80 : 96,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: buttonColor,
              boxShadow: [
                BoxShadow(
                  color: buttonColor.withOpacity(0.3),
                  blurRadius: 20 + pulseValue,
                  spreadRadius: pulseValue,
                ),
              ],
            ),
            child: Center(
              child: widget.isLoading
                  ? const SizedBox(
                      width: 28,
                      height: 28,
                      child: CircularProgressIndicator(
                        color: Colors.white,
                        strokeWidth: 2.5,
                      ),
                    )
                  : AnimatedSwitcher(
                      duration: AppTheme.transitionFast,
                      child: Icon(
                        widget.isRecording ? Icons.stop_rounded : Icons.mic,
                        key: ValueKey(widget.isRecording),
                        size: widget.isRecording ? 32 : 38,
                        color: Colors.white,
                      ),
                    ),
            ),
          );
        },
      ),
    );
  }
}

/// Workaround: AnimatedBuilder is just an alias for AnimatedWidget pattern.
/// Using standard AnimatedBuilder from Flutter.
class AnimatedBuilder extends AnimatedWidget {
  final Widget Function(BuildContext context, Widget? child) builder;
  final Widget? child;

  const AnimatedBuilder({
    super.key,
    required Animation<double> animation,
    required this.builder,
    this.child,
  }) : super(listenable: animation);

  @override
  Widget build(BuildContext context) {
    return builder(context, child);
  }
}
