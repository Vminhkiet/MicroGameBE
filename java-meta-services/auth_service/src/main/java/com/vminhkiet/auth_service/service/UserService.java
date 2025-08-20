package com.vminhkiet.auth_service.service;

import com.vminhkiet.auth_service.dto.AuthResponse;
import com.vminhkiet.auth_service.dto.LoginRequest;
import com.vminhkiet.auth_service.model.User;

import java.util.List;

public interface UserService {
    public AuthResponse loginAccount(LoginRequest request);
    public List<User> getAllUser();
    // public User registerAccount(User request);
    // public User updateAccount(Long Id, User request);
    // public User logoutAccount(User request);
    // public String accessToken(User request);
    // public String refreshToken(User request);
}
