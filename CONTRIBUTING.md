# Contributing to JWT-Crack

Thank you for your interest in contributing to JWT-Crack! This document provides guidelines and information for contributors.

## 🚀 Getting Started

### Prerequisites

- Go 1.22 or higher
- Git
- Basic understanding of JWT tokens and security testing

### Setting up Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/jwt-crack.git
   cd jwt-crack
   ```
3. Add the upstream remote:
   ```bash
   git remote add upstream https://github.com/security-tools/jwt-crack.git
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```
5. Verify setup:
   ```bash
   go build ./cmd/jwt-crack
   go test ./...
   ```

## 📋 Development Guidelines

### Code Style

- Follow standard Go conventions and formatting
- Use `gofmt` to format your code
- Run `golangci-lint` for additional linting
- Write clear, descriptive commit messages

### Architecture Principles

- **Separation of Concerns**: Keep packages focused on single responsibilities
- **Error Handling**: Use custom error types from `internal/errors`
- **Logging**: Use structured logging from `pkg/logger`
- **Validation**: Validate all inputs using `pkg/validator`
- **Testing**: Maintain high test coverage (>80%)

### Package Structure

```
pkg/          # Public, reusable packages
internal/     # Private application code
cmd/          # Application entry points
examples/     # Example configurations and wordlists
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/engine -v

# Run benchmarks
go test -bench=. ./pkg/engine
```

### Writing Tests

- Write unit tests for all new functions
- Include table-driven tests for multiple scenarios
- Test error conditions and edge cases
- Use descriptive test names and comments
- Mock external dependencies when needed

### Test Structure

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name: "valid input",
            input: validInput,
            want: expectedOutput,
            wantErr: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Function() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Function() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## 🔧 Making Changes

### Branch Naming

- `feature/description` - for new features
- `fix/description` - for bug fixes
- `docs/description` - for documentation changes
- `refactor/description` - for code refactoring

### Commit Messages

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: new feature
- `fix`: bug fix
- `docs`: documentation changes
- `style`: formatting, missing semicolons, etc.
- `refactor`: code change that neither fixes bug nor adds feature
- `test`: adding missing tests
- `chore`: changes to build process, auxiliary tools, etc.

Examples:
```
feat(engine): add smart attack pattern recognition
fix(validator): handle malformed JWT tokens correctly
docs(readme): update installation instructions
```

### Pull Request Process

1. Create a feature branch from `main`
2. Make your changes with appropriate tests
3. Update documentation if needed
4. Ensure all tests pass
5. Submit a pull request with:
   - Clear title and description
   - Reference to related issues
   - Screenshots/examples if applicable

## 🐛 Reporting Issues

### Bug Reports

When reporting bugs, please include:

- **Description**: Clear description of the issue
- **Steps to Reproduce**: Detailed steps to reproduce the bug
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**: OS, Go version, JWT-Crack version
- **Sample Data**: Non-sensitive JWT tokens that trigger the issue
- **Logs**: Relevant log output with `--verbose` flag

### Feature Requests

For feature requests, please include:

- **Problem Statement**: What problem does this solve?
- **Proposed Solution**: How would you like it to work?
- **Alternatives**: Other solutions you've considered
- **Use Cases**: Specific scenarios where this would be useful

## 🔒 Security Considerations

### Responsible Disclosure

- Report security vulnerabilities privately to maintainers
- Do not create public issues for security vulnerabilities
- Allow time for fixes before public disclosure

### Code Security

- Never commit sensitive data (tokens, secrets, keys)
- Validate all inputs to prevent injection attacks
- Use secure defaults in configuration
- Follow security best practices in Go

## 📚 Documentation

### Code Documentation

- Use Go doc comments for all public functions and types
- Include usage examples in doc comments
- Document non-obvious behavior and edge cases

### User Documentation

- Update README.md for user-facing changes
- Add examples for new features
- Update help text and command descriptions

## 🎯 Areas for Contribution

### High Priority

- Web interface implementation
- Additional attack algorithms
- Performance optimizations
- Better error messages
- More comprehensive wordlists

### Medium Priority

- Configuration file support
- Plugin system for custom attacks
- Results export formats
- Rate limiting and resource management
- Advanced logging features

### Low Priority

- GUI application
- Docker containerization
- Cloud deployment support
- Integration with other security tools

## 🤝 Community

### Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please:

- Be respectful and professional
- Focus on constructive feedback
- Help others learn and grow
- Report inappropriate behavior

### Getting Help

- **Documentation**: Check README and wiki first
- **Issues**: Search existing issues before creating new ones
- **Discussions**: Use GitHub Discussions for questions
- **Code Review**: Participate in reviewing others' contributions

## 📄 Legal

### Licensing

- All contributions are licensed under MIT License
- You retain copyright to your contributions
- By contributing, you agree to license under project license

### Export Control

This project may be subject to export control regulations. Contributors are responsible for ensuring compliance with applicable laws.

---

Thank you for contributing to JWT-Crack! Your efforts help make security testing more accessible and effective.