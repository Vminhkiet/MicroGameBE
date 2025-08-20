package com.vminhkiet.auth_service.serviceImpl;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.core.userdetails.UserDetails;
import org.springframework.security.core.Authentication;
import org.springframework.stereotype.Service;

import com.vminhkiet.auth_service.config.JwtProvider;
import com.vminhkiet.auth_service.dto.AuthResponse;
import com.vminhkiet.auth_service.dto.LoginRequest;
import com.vminhkiet.auth_service.model.User;
import com.vminhkiet.auth_service.repository.UserRepository;
import com.vminhkiet.auth_service.util.CollectionToStringUtil;

import java.util.List;
import java.util.Map;

@Service
public class UserService implements com.vminhkiet.auth_service.service.UserService{
    @Autowired
    private UserRepository userRepository;
    @Autowired
    private PasswordEncoder passwordEncoder;
    @Autowired
    private CustomUsersDetailService customUsersDetailService;
    @Autowired
    private JwtProvider jwtProvider;
    @Autowired
    private SessionService sessionService;

    @Override
    public AuthResponse loginAccount(LoginRequest request) {
        String userName = request.getUsername();
        String passWord = request.getPassword();

        Authentication auth = authenticate(userName, passWord);

        String roles = CollectionToStringUtil.joinAuthorities(auth.getAuthorities());
        String refreshToken = sessionService.login(Long.parseLong(auth.getName()), roles);

        String jwt = jwtProvider.generateAccessToken(auth.getName(), roles);
                
        AuthResponse authResponse = new AuthResponse();
        authResponse.setJwt(jwt);
        authResponse.setRefreshToken(refreshToken);
        return authResponse;
    }

    private Authentication authenticate(String userName, String passWord) {
        UserDetails userDetails  = customUsersDetailService.loadUserByUsername(userName);

        if (!passwordEncoder.matches(passWord, userDetails.getPassword()))
            throw new BadCredentialsException("Invalid username or password");

        Authentication authObj = new UsernamePasswordAuthenticationToken(userDetails, null, userDetails.getAuthorities());
        return authObj;
    }

    @Override
    public List<User> getAllUser(){
        return userRepository.findAll();
    }
}
