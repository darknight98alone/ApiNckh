# ApiNckh

- 9/10/2019 - completed api push text and id and mac to server to move folder from temp dir to saved folder and push to the database elasticsearch
  json have format:
  {
  "mac":"123456sodp",
  "id":"1",
  "contents":"noi dung cua file"
  }

- API để lưu file cho phép đẩy file lên, trả về text và ID, thêm mức độ để xử lý file đẩy lên, (deskew, deblur, autofill, xử lý bảng,
- [x] API nhận text và ID, đẩy lên elastic search.
- API gửi lên đoạn text, trả về danh sách các tên file và ID đi kèm
- API nhận tên file và ID, chỉ trả về file text
- API Download sẽ tải file gốc về
