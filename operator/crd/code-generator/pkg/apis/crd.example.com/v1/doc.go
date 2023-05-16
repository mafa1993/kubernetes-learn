// 定义在 doc.go 文件的注释，起到的是全局的代码生成控制的作用，所以也被称为 Global Tags。
// +k8s:deepcopy-gen=package
// 定义这个包对应的 API 组的名字
// +groupName=crd.example.com

// Package v1 is the v1 version of the API.
package v1
