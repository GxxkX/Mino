import 'package:flutter/material.dart';
import 'package:permission_handler/permission_handler.dart';
import '../../core/theme/app_theme.dart';
import 'home_screen.dart';

class PermissionScreen extends StatefulWidget {
  const PermissionScreen({super.key});

  @override
  State<PermissionScreen> createState() => _PermissionScreenState();
}

class _PermissionScreenState extends State<PermissionScreen>
    with SingleTickerProviderStateMixin {
  bool _hasPermission = false;
  bool _isChecking = true;
  late AnimationController _animController;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      vsync: this,
      duration: AppTheme.transitionSlow,
    );
    _scaleAnimation = CurvedAnimation(
      parent: _animController,
      curve: Curves.easeOutBack,
    );
    _checkPermission();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  Future<void> _checkPermission() async {
    final status = await Permission.microphone.status;
    setState(() {
      _hasPermission = status.isGranted;
      _isChecking = false;
    });
    _animController.forward();
  }

  Future<void> _requestPermission() async {
    final status = await Permission.microphone.request();
    setState(() {
      _hasPermission = status.isGranted;
    });
  }

  void _onContinue() {
    if (_hasPermission) {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => const HomeScreen()),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.background,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(AppTheme.spaceLg),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              ScaleTransition(
                scale: _scaleAnimation,
                child: Container(
                  width: 96,
                  height: 96,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: (_hasPermission ? AppTheme.cta : AppTheme.secondary)
                        .withOpacity(0.15),
                    boxShadow: _hasPermission
                        ? [
                            BoxShadow(
                              color: AppTheme.cta.withOpacity(0.2),
                              blurRadius: 20,
                              spreadRadius: 2,
                            ),
                          ]
                        : [],
                  ),
                  child: Icon(
                    _hasPermission ? Icons.check_circle : Icons.mic_off,
                    size: 48,
                    color:
                        _hasPermission ? AppTheme.cta : AppTheme.textSecondary,
                  ),
                ),
              ),
              const SizedBox(height: AppTheme.spaceLg),
              Text(
                _hasPermission ? 'Permission Granted' : 'Microphone Access',
                style: AppTheme.headingStyle(fontSize: 24),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: AppTheme.spaceSm),
              Text(
                _hasPermission
                    ? 'You can now use voice recording features.'
                    : 'Mino needs access to your microphone to record voice memos.',
                style: AppTheme.bodyStyle(color: AppTheme.textSecondary),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: AppTheme.spaceXl),
              if (_isChecking)
                const CircularProgressIndicator(color: AppTheme.cta)
              else if (!_hasPermission)
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _requestPermission,
                    child: Text(
                      'Grant Permission',
                      style: AppTheme.bodyStyle(
                        fontWeight: FontWeight.w600,
                        color: Colors.white,
                      ),
                    ),
                  ),
                )
              else
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _onContinue,
                    child: Text(
                      'Continue',
                      style: AppTheme.bodyStyle(
                        fontWeight: FontWeight.w600,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
