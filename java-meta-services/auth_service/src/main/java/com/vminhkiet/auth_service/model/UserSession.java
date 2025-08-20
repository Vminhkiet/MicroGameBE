package com.vminhkiet.auth_service.model;

import java.time.Instant;
import java.util.Set;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import com.vminhkiet.auth_service.model.Role;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class UserSession {
    private String refreshToken;
    private String roles; 
    private Instant loginTime;
}
