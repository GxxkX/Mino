import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../core/theme/app_theme.dart';
import '../../core/providers/auth_provider.dart';
import '../../features/settings/providers/settings_provider.dart';
import 'login_screen.dart';

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  Future<void> _handleSignOut(BuildContext context) async {
    final authProvider = context.read<AuthProvider>();
    await authProvider.signOut();
    if (context.mounted) {
      Navigator.of(context).pushAndRemoveUntil(
        MaterialPageRoute(builder: (_) => const LoginScreen()),
        (route) => false,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<SettingsProvider>(
      builder: (context, settings, _) {
        final t = settings.t;
        return Scaffold(
          backgroundColor: AppTheme.background,
          appBar: AppBar(
            backgroundColor: AppTheme.background,
            elevation: 0,
            title: Text(t('Settings', '设置'),
                style: AppTheme.headingStyle(fontSize: 22)),
            leading: IconButton(
              icon: const Icon(Icons.arrow_back, color: AppTheme.text),
              onPressed: () => Navigator.of(context).pop(),
            ),
          ),
          body: SafeArea(
            child: ListView(
              padding: const EdgeInsets.all(AppTheme.spaceMd),
              children: [
                // Account section
                Consumer<AuthProvider>(
                  builder: (context, auth, _) {
                    return _buildSection(
                      t('ACCOUNT', '账户'),
                      [
                        Padding(
                          padding: const EdgeInsets.symmetric(
                            horizontal: AppTheme.spaceMd,
                            vertical: AppTheme.spaceMd * 0.75,
                          ),
                          child: Row(
                            children: [
                              Container(
                                width: 36,
                                height: 36,
                                decoration: BoxDecoration(
                                  shape: BoxShape.circle,
                                  color: AppTheme.cta.withValues(alpha: 0.15),
                                ),
                                child: Center(
                                  child: Text(
                                    (auth.username ?? 'U')
                                        .substring(0, 1)
                                        .toUpperCase(),
                                    style: AppTheme.bodyStyle(
                                      fontWeight: FontWeight.w600,
                                      color: AppTheme.cta,
                                    ),
                                  ),
                                ),
                              ),
                              const SizedBox(width: AppTheme.spaceMd),
                              Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    auth.username ?? 'User',
                                    style: AppTheme.bodyStyle(
                                        fontWeight: FontWeight.w500),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),
                      ],
                    );
                  },
                ),
                const SizedBox(height: AppTheme.spaceLg),
                // Recording section
                _buildSection(
                  t('RECORDING', '录音'),
                  [
                    Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: AppTheme.spaceMd,
                        vertical: AppTheme.spaceMd * 0.75,
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              const Icon(Icons.mic_outlined,
                                  color: AppTheme.textSecondary, size: 22),
                              const SizedBox(width: AppTheme.spaceMd),
                              Text(t('Recording Gain', '录音增益'),
                                  style: AppTheme.bodyStyle()),
                              const Spacer(),
                              Text(
                                settings.recordingGain == 1.0
                                    ? t('Standard', '标准')
                                    : '${(settings.recordingGain * 100).round()}%',
                                style: AppTheme.bodyStyle(
                                  fontSize: 12,
                                  color: AppTheme.textSecondary,
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: AppTheme.spaceSm),
                          Row(
                            children: [
                              const SizedBox(width: 38),
                              Text(t('Mute', '静音'),
                                  style: AppTheme.bodyStyle(
                                      fontSize: 11,
                                      color: AppTheme.textSecondary)),
                              Expanded(
                                child: Slider(
                                  value: settings.recordingGain,
                                  min: 0.0,
                                  max: 3.0,
                                  divisions: 30,
                                  activeColor: AppTheme.cta,
                                  inactiveColor: AppTheme.secondary,
                                  onChanged: (v) =>
                                      settings.setRecordingGain(v),
                                ),
                              ),
                              Text('3x',
                                  style: AppTheme.bodyStyle(
                                      fontSize: 11,
                                      color: AppTheme.textSecondary)),
                            ],
                          ),
                          if (settings.recordingGain > 1.0)
                            Padding(
                              padding: const EdgeInsets.only(
                                  left: 38, top: AppTheme.spaceXs),
                              child: Text(
                                t('Gain above 100% may cause audio clipping',
                                    '增益大于 100% 时录音可能产生削波失真'),
                                style: AppTheme.bodyStyle(
                                  fontSize: 11,
                                  color: AppTheme.warning,
                                ),
                              ),
                            ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: AppTheme.spaceLg),
                // General section
                _buildSection(
                  t('GENERAL', '通用'),
                  [
                    _buildTile(
                      icon: Icons.language,
                      title: t('Language', '语言'),
                      trailing: DropdownButtonHideUnderline(
                        child: DropdownButton<String>(
                          value: settings.language,
                          dropdownColor: AppTheme.primary,
                          style: AppTheme.bodyStyle(fontSize: 13),
                          items: const [
                            DropdownMenuItem(
                                value: 'zh-CN', child: Text('中文')),
                            DropdownMenuItem(
                                value: 'en-US', child: Text('English')),
                          ],
                          onChanged: (v) {
                            if (v != null) settings.setLanguage(v);
                          },
                        ),
                      ),
                    ),
                    const Divider(height: 0.5, indent: 56),
                    _buildTile(
                      icon: Icons.info_outline,
                      title: t('Version', '版本'),
                      subtitle: '1.0.0',
                    ),
                  ],
                ),
                const SizedBox(height: AppTheme.space2xl),
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton(
                    onPressed: () => _handleSignOut(context),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: AppTheme.error,
                      side: const BorderSide(color: AppTheme.error),
                      padding: const EdgeInsets.symmetric(vertical: 12),
                    ),
                    child: Text(
                      t('Sign Out', '退出登录'),
                      style: AppTheme.bodyStyle(
                        fontWeight: FontWeight.w600,
                        color: AppTheme.error,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildSection(String title, List<Widget> children) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(
              left: AppTheme.spaceSm, bottom: AppTheme.spaceSm),
          child: Text(
            title,
            style: AppTheme.bodyStyle(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: AppTheme.textSecondary,
            ),
          ),
        ),
        Container(
          decoration: BoxDecoration(
            color: AppTheme.primary,
            borderRadius: BorderRadius.circular(AppTheme.radiusMd),
            boxShadow: AppTheme.shadowSm,
          ),
          clipBehavior: Clip.antiAlias,
          child: Column(children: children),
        ),
      ],
    );
  }

  Widget _buildTile({
    required IconData icon,
    required String title,
    String? subtitle,
    Widget? trailing,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(
        horizontal: AppTheme.spaceMd,
        vertical: AppTheme.spaceMd * 0.75,
      ),
      child: Row(
        children: [
          Icon(icon, color: AppTheme.textSecondary, size: 22),
          const SizedBox(width: AppTheme.spaceMd),
          Expanded(child: Text(title, style: AppTheme.bodyStyle())),
          if (trailing != null)
            trailing
          else if (subtitle != null)
            Text(
              subtitle,
              style: AppTheme.bodyStyle(
                  fontSize: 12, color: AppTheme.textSecondary),
            ),
        ],
      ),
    );
  }
}
