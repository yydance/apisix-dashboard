/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package log

var (
	DefLogger Interface = emptyLog{}
)

type Type int8

const (
	AccessLog Type = iota - 1
	ErrorLog
)

type emptyLog struct {
}

type Interface interface {
	Debug(msg string, fields ...any)
	Debugf(msg string, args ...any)
	Info(msg string, fields ...any)
	Infof(msg string, args ...any)
	Warn(msg string, fields ...any)
	Warnf(msg string, args ...any)
	Error(msg string, fields ...any)
	Errorf(msg string, args ...any)
	Fatal(msg string, fields ...any)
	Fatalf(msg string, args ...any)
}

func (e emptyLog) Debug(msg string, fields ...any) {
	getZapFields(logger, fields).Debug(msg)
}

func (e emptyLog) Debugf(msg string, args ...any) {
	logger.Debugf(msg, args...)
}

func (e emptyLog) Info(msg string, fields ...any) {
	getZapFields(logger, fields).Info(msg)
}

func (e emptyLog) Infof(msg string, args ...any) {
	logger.Infof(msg, args...)
}

func (e emptyLog) Warn(msg string, fields ...any) {
	getZapFields(logger, fields).Warn(msg)
}

func (e emptyLog) Warnf(msg string, args ...any) {
	logger.Warnf(msg, args...)
}

func (e emptyLog) Error(msg string, fields ...any) {
	getZapFields(logger, fields).Error(msg)
}

func (e emptyLog) Errorf(msg string, args ...any) {
	logger.Errorf(msg, args...)
}

func (e emptyLog) Fatal(msg string, fields ...any) {
	getZapFields(logger, fields).Fatal(msg)
}

func (e emptyLog) Fatalf(msg string, args ...any) {
	logger.Fatalf(msg, args...)
}

func Debug(msg string, fields ...any) {
	DefLogger.Debug(msg, fields...)
}
func Debugf(msg string, args ...any) {
	DefLogger.Debugf(msg, args...)
}
func Info(msg string, fields ...any) {
	DefLogger.Info(msg, fields...)
}
func Infof(msg string, args ...any) {
	DefLogger.Infof(msg, args...)
}
func Warn(msg string, fields ...any) {
	DefLogger.Warn(msg, fields...)
}
func Warnf(msg string, args ...any) {
	DefLogger.Warnf(msg, args...)
}
func Error(msg string, fields ...any) {
	DefLogger.Error(msg, fields...)
}
func Errorf(msg string, args ...any) {
	DefLogger.Errorf(msg, args...)
}
func Fatal(msg string, fields ...any) {
	DefLogger.Fatal(msg, fields...)
}
func Fatalf(msg string, args ...any) {
	DefLogger.Fatalf(msg, args...)
}
