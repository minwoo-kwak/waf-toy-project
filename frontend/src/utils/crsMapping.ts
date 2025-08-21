// OWASP CRS (Core Rule Set) 공격 유형 매핑 테이블
// CRS 규칙 ID를 실제 공격 유형으로 변환하는 유틸리티

export interface AttackTypeInfo {
  category: string;
  severity: 'HIGH' | 'MEDIUM' | 'LOW';
  description: string;
  color: string;
  icon: string;
}

// CRS 룰 ID별 공격 유형 매핑
export const CRS_ATTACK_MAPPING: Record<number, AttackTypeInfo> = {
  // SQL Injection 공격 (920xxx-921xxx 범위)
  920001: {
    category: 'SQL Injection',
    severity: 'HIGH',
    description: 'SQL injection attack detected',
    color: '#ff4757',
    icon: '💉'
  },
  920002: {
    category: 'SQL Injection',
    severity: 'HIGH', 
    description: 'SQL injection - tautology attack',
    color: '#ff4757',
    icon: '💉'
  },
  920003: {
    category: 'SQL Injection',
    severity: 'HIGH',
    description: 'SQL injection - union attack',
    color: '#ff4757',
    icon: '💉'
  },

  // XSS 공격 (941xxx 범위)
  941100: {
    category: 'XSS (Cross-Site Scripting)',
    severity: 'HIGH',
    description: 'XSS attack detected',
    color: '#ff6b35',
    icon: '🔗'
  },
  941110: {
    category: 'XSS (Cross-Site Scripting)',
    severity: 'HIGH',
    description: 'XSS filter - script tag attack',
    color: '#ff6b35',
    icon: '🔗'
  },
  941120: {
    category: 'XSS (Cross-Site Scripting)',
    severity: 'HIGH',
    description: 'XSS filter - event handler attack',
    color: '#ff6b35',
    icon: '🔗'
  },

  // Command Injection (932xxx 범위)
  932100: {
    category: 'Command Injection',
    severity: 'HIGH',
    description: 'Remote command execution attack',
    color: '#e55039',
    icon: '💻'
  },
  932110: {
    category: 'Command Injection',
    severity: 'HIGH',
    description: 'Unix command injection',
    color: '#e55039',
    icon: '💻'
  },
  932120: {
    category: 'Command Injection',
    severity: 'HIGH',
    description: 'Windows command injection',
    color: '#e55039',
    icon: '💻'
  },

  // Path Traversal (930xxx 범위)
  930100: {
    category: 'Path Traversal',
    severity: 'MEDIUM',
    description: 'Path traversal attack detected',
    color: '#ffa502',
    icon: '📁'
  },
  930110: {
    category: 'Path Traversal',
    severity: 'MEDIUM',
    description: 'Directory traversal attack',
    color: '#ffa502',
    icon: '📁'
  },

  // LFI/RFI (931xxx 범위)
  931100: {
    category: 'Local File Inclusion',
    severity: 'HIGH',
    description: 'Local file inclusion attack',
    color: '#ff7675',
    icon: '📂'
  },
  931110: {
    category: 'Remote File Inclusion',
    severity: 'HIGH',
    description: 'Remote file inclusion attack',
    color: '#fd79a8',
    icon: '🌐'
  },

  // HTTP Protocol Violations (920xxx 범위)
  920100: {
    category: 'Protocol Violation',
    severity: 'MEDIUM',
    description: 'Invalid HTTP request',
    color: '#6c5ce7',
    icon: '⚠️'
  },
  920200: {
    category: 'Protocol Violation', 
    severity: 'MEDIUM',
    description: 'Range header abuse',
    color: '#6c5ce7',
    icon: '⚠️'
  },

  // Scanner Detection (913xxx 범위)
  913100: {
    category: 'Scanner Detection',
    severity: 'LOW',
    description: 'Security scanner detected',
    color: '#74b9ff',
    icon: '🔍'
  },
  913110: {
    category: 'Scanner Detection',
    severity: 'LOW',
    description: 'Automated tool detected',
    color: '#74b9ff',
    icon: '🤖'
  },

  // Generic Attacks
  949110: {
    category: 'Generic Attack',
    severity: 'MEDIUM',
    description: 'Inbound anomaly score exceeded',
    color: '#a29bfe',
    icon: '🚨'
  },

  // Rate Limiting
  912001: {
    category: 'Rate Limiting',
    severity: 'LOW',
    description: 'Request rate limit exceeded',
    color: '#00cec9',
    icon: '⏱️'
  }
};

// CRS 룰 범위별 카테고리 매핑
export const CRS_RANGE_MAPPING: Record<string, AttackTypeInfo> = {
  '920': {
    category: 'Protocol Violation',
    severity: 'MEDIUM',
    description: 'HTTP protocol attack or violation',
    color: '#6c5ce7',
    icon: '⚠️'
  },
  '921': {
    category: 'Protocol Anomaly', 
    severity: 'MEDIUM',
    description: 'HTTP protocol anomaly',
    color: '#fd79a8',
    icon: '📊'
  },
  '930': {
    category: 'Application Attack',
    severity: 'HIGH',
    description: 'Application layer attack',
    color: '#fdcb6e',
    icon: '🎯'
  },
  '931': {
    category: 'Application Attack',
    severity: 'HIGH', 
    description: 'Application layer attack',
    color: '#fdcb6e',
    icon: '🎯'
  },
  '932': {
    category: 'Application Attack',
    severity: 'HIGH',
    description: 'Application layer attack', 
    color: '#fdcb6e',
    icon: '🎯'
  },
  '933': {
    category: 'Application Attack',
    severity: 'HIGH',
    description: 'Application layer attack',
    color: '#fdcb6e', 
    icon: '🎯'
  },
  '941': {
    category: 'XSS (Cross-Site Scripting)',
    severity: 'HIGH',
    description: 'Cross-site scripting attack',
    color: '#ff6b35',
    icon: '🔗'
  },
  '942': {
    category: 'SQL Injection',
    severity: 'HIGH',
    description: 'SQL injection attack',
    color: '#ff4757',
    icon: '💉'
  },
  '943': {
    category: 'Session Fixation',
    severity: 'MEDIUM',
    description: 'Session fixation attack',
    color: '#55a3ff',
    icon: '🔐'
  },
  '944': {
    category: 'Java Attack',
    severity: 'HIGH', 
    description: 'Java application attack',
    color: '#ff9ff3',
    icon: '☕'
  }
};

/**
 * CRS 룰 ID를 기반으로 공격 유형 정보를 반환
 */
export function getAttackTypeFromRuleId(ruleId: number | string): AttackTypeInfo {
  const numericRuleId = typeof ruleId === 'string' ? parseInt(ruleId) : ruleId;
  
  // 정확한 룰 ID 매칭
  if (CRS_ATTACK_MAPPING[numericRuleId]) {
    return CRS_ATTACK_MAPPING[numericRuleId];
  }
  
  // 룰 범위로 매칭 (예: 942xxx -> SQL Injection)
  const ruleIdStr = numericRuleId.toString();
  const rangePrefix = ruleIdStr.substring(0, 3);
  
  if (CRS_RANGE_MAPPING[rangePrefix]) {
    return CRS_RANGE_MAPPING[rangePrefix];
  }
  
  // 기본값 반환
  return {
    category: 'Unknown Attack',
    severity: 'MEDIUM',
    description: `Unknown attack pattern (Rule ID: ${numericRuleId})`,
    color: '#636e72',
    icon: '❓'
  };
}

/**
 * 로그 메시지에서 CRS 룰 ID 추출
 */
export function extractRuleIdsFromMessage(message: string): number[] {
  const ruleIdPattern = /\[id "(\d+)"\]/g;
  const matches = [...message.matchAll(ruleIdPattern)];
  return matches.map(match => parseInt(match[1])).filter(id => !isNaN(id));
}

/**
 * 여러 룰 ID에서 가장 심각한 공격 유형 반환
 */
export function getMostSevereAttackType(ruleIds: number[]): AttackTypeInfo {
  if (ruleIds.length === 0) {
    return getAttackTypeFromRuleId(0); // 기본값
  }
  
  const attackTypes = ruleIds.map(getAttackTypeFromRuleId);
  const severityOrder = { 'HIGH': 3, 'MEDIUM': 2, 'LOW': 1 };
  
  return attackTypes.reduce((mostSevere, current) => {
    return severityOrder[current.severity] > severityOrder[mostSevere.severity] 
      ? current : mostSevere;
  });
}

/**
 * 공격 유형별 통계 집계
 */
export function aggregateAttackStats(logs: any[]): Record<string, number> {
  const stats: Record<string, number> = {};
  
  logs.forEach(log => {
    if (log.message) {
      const ruleIds = extractRuleIdsFromMessage(log.message);
      const attackType = ruleIds.length > 0 
        ? getMostSevereAttackType(ruleIds)
        : getAttackTypeFromRuleId(0);
      
      stats[attackType.category] = (stats[attackType.category] || 0) + 1;
    }
  });
  
  return stats;
}