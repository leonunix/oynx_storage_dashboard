const ERROR_KEY_MAP = {
  'not found': 'errors.notFound',
  'invalid request body': 'errors.invalidRequestBody',
  'dashboard setup required': 'errors.setupRequired',
  'invalid credentials': 'errors.invalidCredentials',
  'failed to issue token': 'errors.failedToIssueToken',
  'missing bearer token': 'errors.missingBearerToken',
  'invalid token': 'errors.invalidToken',
  'unauthenticated': 'errors.unauthenticated',
  'permission denied': 'errors.permissionDenied',
  'user disabled': 'errors.userDisabled',
  'username is required': 'errors.usernameRequired',
  'password is required': 'errors.passwordRequired',
  'dashboard already initialized': 'errors.alreadyInitialized',
}

export function translateError(errorString, t) {
  if (!errorString) return t('errors.unknown')
  const key = ERROR_KEY_MAP[errorString.toLowerCase()]
  if (key) return t(key)
  return errorString
}
