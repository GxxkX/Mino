import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../core/theme/app_theme.dart';
import '../../core/constants/app_config.dart';
import '../../core/providers/auth_provider.dart';
import '../../core/services/auth_service.dart';
import 'home_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen>
    with SingleTickerProviderStateMixin {
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _apiUrlController = TextEditingController();
  bool _obscurePassword = true;
  bool _showApiUrlField = false;
  late AnimationController _fadeController;
  late Animation<double> _fadeAnimation;

  @override
  void initState() {
    super.initState();
    _fadeController = AnimationController(
      vsync: this,
      duration: AppTheme.transitionSlow,
    );
    _fadeAnimation = CurvedAnimation(
      parent: _fadeController,
      curve: Curves.easeOut,
    );
    _fadeController.forward();
    _apiUrlController.text = AppConfig.apiUrl;
    _loadSavedCredentials();
  }

  Future<void> _loadSavedCredentials() async {
    final creds = await AuthService().getSavedCredentials();
    if (!mounted) return;
    if (creds['username'] != null) {
      _usernameController.text = creds['username']!;
    }
    if (creds['password'] != null) {
      _passwordController.text = creds['password']!;
    }
    if (creds['apiUrl'] != null) {
      _apiUrlController.text = creds['apiUrl']!;
    }
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    _apiUrlController.dispose();
    _fadeController.dispose();
    super.dispose();
  }

  Future<void> _handleLogin() async {
    final authProvider = context.read<AuthProvider>();
    final username = _usernameController.text.trim();
    final password = _passwordController.text;
    final apiUrl = _apiUrlController.text.trim();

    if (username.isEmpty || password.isEmpty) {
      return;
    }

    if (apiUrl.isNotEmpty) {
      await AppConfig.setApiUrl(apiUrl);
    }

    final success = await authProvider.signIn(username, password);

    if (success && mounted) {
      await AuthService().saveCredentials(
          username, password, apiUrl.isNotEmpty ? apiUrl : AppConfig.apiUrl);
      if (!mounted) return;
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
        child: FadeTransition(
          opacity: _fadeAnimation,
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(AppTheme.spaceLg),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const SizedBox(height: AppTheme.space2xl),
                // Logo
                Container(
                  width: 80,
                  height: 80,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    boxShadow: [
                      BoxShadow(
                        color: AppTheme.cta.withOpacity(0.2),
                        blurRadius: 24,
                        spreadRadius: 4,
                      ),
                    ],
                  ),
                  child: ClipOval(
                    child: Image.asset(
                      'assets/logo.png',
                      width: 80,
                      height: 80,
                      fit: BoxFit.cover,
                    ),
                  ),
                ),
                const SizedBox(height: AppTheme.spaceMd),
                Text(
                  'Mino',
                  style: AppTheme.headingStyle(fontSize: 36),
                ),
                const SizedBox(height: AppTheme.spaceXs),
                Text(
                  'Your AI Assistant',
                  style: AppTheme.bodyStyle(
                    color: AppTheme.textSecondary,
                  ),
                ),
                const SizedBox(height: AppTheme.space2xl),
                // Username field
                TextField(
                  controller: _usernameController,
                  style: AppTheme.bodyStyle(),
                  decoration: const InputDecoration(
                    hintText: 'Username',
                    prefixIcon: Icon(
                      Icons.person_outline,
                      color: AppTheme.textSecondary,
                    ),
                  ),
                  textInputAction: TextInputAction.next,
                ),
                const SizedBox(height: AppTheme.spaceMd),
                // Password field
                TextField(
                  controller: _passwordController,
                  style: AppTheme.bodyStyle(),
                  obscureText: _obscurePassword,
                  decoration: InputDecoration(
                    hintText: 'Password',
                    prefixIcon: const Icon(
                      Icons.lock_outline,
                      color: AppTheme.textSecondary,
                    ),
                    suffixIcon: IconButton(
                      icon: Icon(
                        _obscurePassword
                            ? Icons.visibility_off
                            : Icons.visibility,
                        color: AppTheme.textSecondary,
                      ),
                      onPressed: () {
                        setState(() {
                          _obscurePassword = !_obscurePassword;
                        });
                      },
                    ),
                  ),
                  textInputAction: TextInputAction.done,
                  onSubmitted: (_) => _handleLogin(),
                ),
                const SizedBox(height: AppTheme.spaceMd),
                // API URL field (collapsible)
                AnimatedContainer(
                  duration: AppTheme.transitionFast,
                  height: _showApiUrlField ? 60 : 0,
                  child: _showApiUrlField
                      ? TextField(
                          controller: _apiUrlController,
                          style: AppTheme.bodyStyle(),
                          decoration: const InputDecoration(
                            hintText: 'API URL',
                            prefixIcon: Icon(
                              Icons.link,
                              color: AppTheme.textSecondary,
                            ),
                          ),
                          textInputAction: TextInputAction.done,
                          onSubmitted: (_) => _handleLogin(),
                        )
                      : const SizedBox.shrink(),
                ),
                if (_showApiUrlField)
                  const SizedBox(height: AppTheme.spaceMd),
                // Toggle API URL button
                TextButton.icon(
                  onPressed: () {
                    setState(() {
                      _showApiUrlField = !_showApiUrlField;
                    });
                  },
                  icon: Icon(
                    _showApiUrlField
                        ? Icons.expand_less
                        : Icons.settings_outlined,
                    size: 18,
                    color: AppTheme.textSecondary,
                  ),
                  label: Text(
                    _showApiUrlField ? 'Hide API Settings' : 'Show API Settings',
                    style: AppTheme.bodyStyle(
                      fontSize: 12,
                      color: AppTheme.textSecondary,
                    ),
                  ),
                ),
                const SizedBox(height: AppTheme.spaceLg),
                // Sign in button
                Consumer<AuthProvider>(
                  builder: (context, auth, _) {
                    return SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: auth.isLoading ? null : _handleLogin,
                        child: AnimatedSwitcher(
                          duration: AppTheme.transitionFast,
                          child: auth.isLoading
                              ? const SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    color: Colors.white,
                                  ),
                                )
                              : Text(
                                  'Sign In',
                                  style: AppTheme.bodyStyle(
                                    fontWeight: FontWeight.w600,
                                    color: Colors.white,
                                  ),
                                ),
                        ),
                      ),
                    );
                  },
                ),
                // Error message
                Consumer<AuthProvider>(
                  builder: (context, auth, _) {
                    if (auth.error == null) return const SizedBox.shrink();
                    return Padding(
                      padding: const EdgeInsets.only(top: AppTheme.spaceMd),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: AppTheme.spaceMd,
                          vertical: AppTheme.spaceSm,
                        ),
                        decoration: BoxDecoration(
                          color: AppTheme.error.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(
                            AppTheme.radiusSm,
                          ),
                        ),
                        child: Text(
                          auth.error!,
                          style: AppTheme.bodyStyle(
                            fontSize: 12,
                            color: AppTheme.error,
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
