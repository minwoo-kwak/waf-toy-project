export interface WAFLog {
  id: string;
  timestamp: string;
  client_ip: string;
  method: string;
  url: string;
  user_agent: string;
  attack_type: string;
  rule_id: string;
  message: string;
  blocked: boolean;
  severity: string;
  raw_log: string;
}

export interface IPStat {
  ip: string;
  requests: number;
  blocked: number;
}

export interface WAFStats {
  total_requests: number;
  blocked_requests: number;
  attacks_by_type: Record<string, number>;
  top_ips: IPStat[];
  recent_logs: WAFLog[];
  timestamp: string;
}

export interface CustomRule {
  id: string;
  name: string;
  description: string;
  rule_text: string;
  enabled: boolean;
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
  created_at: string;
  updated_at: string;
}

export interface CustomRuleRequest {
  name: string;
  description: string;
  rule_text: string;
  enabled: boolean;
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';
}

export interface SecurityTest {
  id: string;
  name: string;
  description: string;
  test_type: string;
  payloads: string[];
  results: SecurityResult[];
  created_at: string;
}

export interface SecurityResult {
  payload: string;
  blocked: boolean;
  status_code: number;
  response: string;
}

export interface SecurityTestRequest {
  test_type: 'sql_injection' | 'xss' | 'path_traversal' | 'command_injection';
  payloads?: string[];
}