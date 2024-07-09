// Copyright 2000-2024 JetBrains s.r.o. and contributors. Use of this source code is governed by the Apache 2.0 license.

package org.jetbrains.intellij.platform.gradle

/**
 * Represents an annotation to mark DSL elements related to the IntelliJ Platform.
 */
@DslMarker
//@RequiresOptIn(message = "This API belongs to ${Plugin.ID}")
annotation class IntelliJPlatform
