# ApiNckh  
- API để lưu file cho phép đẩy file lên, trả về text và ID, thêm mức độ để xử lý file đẩy lên, (deskew, deblur, autofill, xử lý bảng,
- [x] API nhận text và ID, đẩy lên elastic search.
json client gửi lên có định dạng:  
  {  
  &nbsp;&nbsp;&nbsp;"mac":"123456sodp",  
  &nbsp;&nbsp;&nbsp;"id":"1",  
  &nbsp;&nbsp;&nbsp;"contents":"noi dung cua file"  
  }  
- [x] API gửi lên đoạn text, trả về danh sách các tên file, nội dung của file được giới hạn 100 từ và ID đi kèm
json client gửi lên có dạng:  
{  
	&nbsp;&nbsp;&nbsp;"mac":"documents",  
	&nbsp;&nbsp;&nbsp;"search_contents":"khong co viec gi kho"  
}  
json trả về có dạng mảng:  
[
    {  
        &nbsp;&nbsp;&nbsp;"_source": {  
            &nbsp;&nbsp;&nbsp;"id": "1",  
            &nbsp;&nbsp;&nbsp;"contents": "khong co",  
            &nbsp;&nbsp;&nbsp;"filename": "test.txt"  
        &nbsp;&nbsp;&nbsp;}  
    },  
    {  
        &nbsp;&nbsp;&nbsp;"_source": {  
            &nbsp;&nbsp;&nbsp;"id": "2",  
            &nbsp;&nbsp;&nbsp;"contents": "không có việc gì khó chỉ sợ lòng không bền",  
            &nbsp;&nbsp;&nbsp;"filename": "temp.txt"  
        &nbsp;&nbsp;&nbsp;}  
    }
]  
- API nhận ID, chỉ trả về file text
- API Download sẽ tải file gốc về
