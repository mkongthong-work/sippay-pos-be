package utils

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ไฟล์นี้รวมช่องทางอัปโหลดไฟล์ (รูปสลิปโอนเงิน / รูปเมนู) ไว้ที่เดียว แล้วสลับพฤติกรรมตาม env:
//   - ถ้าตั้ง SUPABASE_URL + SUPABASE_SERVICE_ROLE_KEY + SUPABASE_STORAGE_BUCKET ไว้ครบ
//     → อัปโหลดขึ้น Supabase Storage แล้วคืน public URL เต็ม (จำเป็นตอน deploy บน Vercel เพราะ
//       serverless function ไม่มีดิสก์ถาวรให้เก็บไฟล์ — ทุกครั้งที่ฟังก์ชันจบการทำงาน ไฟล์ที่เขียนไว้จะหายไป)
//   - ถ้าไม่ได้ตั้ง (เช่นรัน backend เองบนเครื่อง/VM ปกติ) → เขียนลงดิสก์ในโฟลเดอร์ ./uploads/<folder>
//     เหมือนเดิม คืน path สัมพัทธ์ (/uploads/<folder>/<ชื่อไฟล์>) แบบเดิมทุกประการ ไม่กระทบ dev workflow เดิม

// SaveUpload บันทึกไฟล์ที่อัปโหลดมา (multipart) ไว้ในโฟลเดอร์ย่อยชื่อ folder (เช่น "slips", "menu-items")
// คืนค่า URL ที่ใช้เปิดดูไฟล์ได้ — อาจเป็น URL เต็ม (Supabase) หรือ path สัมพัทธ์ (เก็บบนดิสก์เอง) แล้วแต่โหมด
func SaveUpload(file *multipart.FileHeader, folder, filenamePrefix string) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%s_%d%s", filenamePrefix, time.Now().UnixNano(), ext)

	if useSupabaseStorage() {
		return uploadToSupabaseStorage(file, folder, filename)
	}
	return saveToLocalDisk(file, folder, filename)
}

func useSupabaseStorage() bool {
	return os.Getenv("SUPABASE_URL") != "" &&
		os.Getenv("SUPABASE_SERVICE_ROLE_KEY") != "" &&
		os.Getenv("SUPABASE_STORAGE_BUCKET") != ""
}

func saveToLocalDisk(file *multipart.FileHeader, folder, filename string) (string, error) {
	dir := filepath.Join("./uploads", folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("สร้างโฟลเดอร์เก็บไฟล์ไม่สำเร็จ: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fullPath := filepath.Join(dir, filename)
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return "/uploads/" + folder + "/" + filename, nil
}

// uploadToSupabaseStorage อัปโหลดไฟล์ขึ้น Supabase Storage ผ่าน REST API ตรงๆ (ไม่ใช้ SDK แยก
// กันเพิ่ม dependency โดยไม่จำเป็น — endpoint นี้เสถียรและเอกสารสาธารณะของ Supabase)
// ต้องตั้งค่า bucket เป็น public bucket ไว้ก่อน (ทำได้จาก Supabase Dashboard > Storage) ไม่งั้น URL
// ที่คืนกลับไปจะเปิดไม่ได้เพราะต้องแนบ auth token ทุกครั้งที่ขอดูไฟล์
func uploadToSupabaseStorage(file *multipart.FileHeader, folder, filename string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, src); err != nil {
		return "", err
	}

	baseURL := strings.TrimRight(os.Getenv("SUPABASE_URL"), "/")
	bucket := os.Getenv("SUPABASE_STORAGE_BUCKET")
	serviceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	objectPath := folder + "/" + filename

	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", baseURL, bucket, objectPath)
	req, err := http.NewRequest(http.MethodPost, uploadURL, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", err
	}
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("apikey", serviceKey)
	req.Header.Set("x-upsert", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("อัปโหลดขึ้น Supabase Storage ไม่สำเร็จ: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Supabase Storage ตอบกลับ %d: %s", resp.StatusCode, string(body))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", baseURL, bucket, objectPath)
	return publicURL, nil
}
